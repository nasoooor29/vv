package services

import (
	"database/sql"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strings"
	"time"

	"visory/internal/database"
	dbsessions "visory/internal/database/sessions"
	"visory/internal/database/user"
	"visory/internal/models"
	"visory/internal/utils"

	"github.com/gofrs/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db             *database.Service
	dispatcher     *utils.Dispatcher
	OAuthProviders map[string]goth.Provider
}

func (s *AuthService) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie(models.COOKIE_NAME)
		if err != nil {
			return s.dispatcher.NewUnauthorized("Failed to get user by session token", err)
		}
		userWithSession, err := s.db.User.GetUserAndSessionByToken(c.Request().Context(), cookie.Value)
		if err != nil {
			return s.dispatcher.NewUnauthorized("Failed to get user by session token", err)
		}
		// NOTE: check if session is expired
		c.Set("userWithSession", userWithSession)

		return next(c)
	}
}

func (s *AuthService) RBACMiddleware(policies ...models.RBACPolicy) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(models.COOKIE_NAME)
			if err != nil {
				return s.dispatcher.NewUnauthorized("Failed to get user by session token", err)
			}
			user, err := s.db.User.GetBySessionToken(c.Request().Context(), cookie.Value)
			if err != nil {
				return s.dispatcher.NewUnauthorized("Failed to get user by session token", err)
			}

			user_roles := models.RoleToRBACPolicies(user.Role)
			if _, ok := user_roles[models.RBAC_USER_ADMIN]; ok {
				return next(c)
			}
			for _, policy := range policies {
				if v, ok := user_roles[policy]; !ok || !v {
					return s.dispatcher.NewForbidden("Insufficient permissions", err)
				}
			}

			return next(c)
		}
	}
}

// NewAuthService creates a new AuthService with dependency injection
func NewAuthService(db *database.Service, dispatcher *utils.Dispatcher) *AuthService {
	// Create a grouped logger for auth service
	authDispatcher := dispatcher.WithGroup("auth")
	authService := &AuthService{
		db:         db,
		dispatcher: authDispatcher,
	}
	providers := authService.initializeOAuth()
	authService.OAuthProviders = providers
	return authService
}

// initializeOAuth sets up OAuth providers with environment variables
func (s *AuthService) initializeOAuth() map[string]goth.Provider {
	gothic.Store = sessions.NewCookieStore([]byte(models.ENV_VARS.SessionSecret))
	baseCallbackURL := models.ENV_VARS.BaseUrlWithPort + "/api/auth/oauth/callback"

	providersMap := map[string]goth.Provider{}
	googleProvider := google.New(
		models.ENV_VARS.GoogleOAuthKey,
		models.ENV_VARS.GoogleOAuthSecret,
		baseCallbackURL+"/google",
		"email", "profile",
	)
	providersMap[googleProvider.Name()] = googleProvider
	githubProvider := github.New(
		models.ENV_VARS.GithubOAuthKey,
		models.ENV_VARS.GithubOAuthSecret,
		baseCallbackURL+"/github",
		"user:email",
	)
	providersMap[githubProvider.Name()] = githubProvider
	providers := slices.Collect(maps.Values(providersMap))
	for _, provider := range providers {
		slog.Info("OAuth provider initialized", "provider", provider.Name())
	}

	goth.UseProviders(providers...)
	return providersMap
}

// Me returns the current authenticated user
func (s *AuthService) Me(c echo.Context) error {
	cookie, err := c.Cookie(models.COOKIE_NAME)
	if err != nil {
		return s.dispatcher.NewUnauthorized("Failed to get user by session token", err)
	}
	userWithSession, err := s.db.User.GetUserAndSessionByToken(c.Request().Context(), cookie.Value)
	if err != nil {
		return s.dispatcher.NewUnauthorized("Failed to get user by session token", err)
	}
	return c.JSON(http.StatusOK, userWithSession)
}

// Logout logs out the current user
func (s *AuthService) Logout(c echo.Context) error {
	cookie, err := c.Cookie(models.COOKIE_NAME)
	if err != nil {
		return s.dispatcher.NewBadRequest("Invalid cookie", err)
	}

	if err := s.db.Session.DeleteBySessionToken(c.Request().Context(), cookie.Value); err != nil {
		return s.dispatcher.NewInternalServerError("Failed to logout", err)
	}

	cookie.MaxAge = -1 // Expire the cookie
	c.SetCookie(cookie)

	return c.NoContent(http.StatusOK)
}

// Register registers a new user
func (s *AuthService) Register(c echo.Context) error {
	p := user.UpsertUserParams{}
	if err := c.Bind(&p); err != nil {
		return s.dispatcher.NewBadRequest("Invalid request body", err)
	}
	_, err := s.db.User.GetByEmailOrUsername(c.Request().Context(), user.GetByEmailOrUsernameParams{
		Email:    p.Email,
		Username: p.Username,
	})
	if err == nil {
		return s.dispatcher.NewConflict("User already exists", nil)
	}
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		return s.dispatcher.NewInternalServerError("Failed to hash password", err)
	}
	p.Password = string(bcryptPassword)
	val, err := s.db.User.UpsertUser(c.Request().Context(), p)
	if err != nil {
		return s.dispatcher.NewInternalServerError("Failed to register user", err)
	}
	if err := s.generateCookie(c, val.ID); err != nil {
		return s.dispatcher.NewInternalServerError("Failed to generate cookie", err)
	}

	return c.JSON(http.StatusOK, val)
}

// Login logs in a user
func (s *AuthService) Login(c echo.Context) error {
	p := models.Login{}
	if err := c.Bind(&p); err != nil {
		return s.dispatcher.NewBadRequest("Invalid request body", err)
	}
	val, err := s.db.User.GetByEmailOrUsername(c.Request().Context(), user.GetByEmailOrUsernameParams{
		Email:    p.Username,
		Username: p.Username,
	})
	if err == sql.ErrNoRows {
		return s.dispatcher.NewNotFound("You don't have an account please register", err)
	}
	if err != nil {
		return s.dispatcher.NewUnauthorized("Failed to login user", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(val.Password), []byte(p.Password))
	if err != nil {
		return s.dispatcher.NewUnauthorized("your username or password is wrong", err)
	}

	if err := s.generateCookie(c, val.ID); err != nil {
		return s.dispatcher.NewInternalServerError("Failed to generate cookie", err)
	}

	return c.JSON(http.StatusOK, val)
}

// OAuthLogin handles OAuth login initiation
func (s *AuthService) OAuthLogin(c echo.Context) error {
	provider := c.Param("provider")
	_, ok := s.OAuthProviders[provider]
	if !ok {
		return s.dispatcher.NewBadRequest("Invalid OAuth provider", fmt.Errorf("provider %s not supported", provider))
	}

	// Add provider context to the request
	req := gothic.GetContextWithProvider(c.Request(), provider)
	c.SetRequest(req)

	// Get the OAuth URL using gothic
	authURL, err := gothic.GetAuthURL(c.Response(), c.Request())
	if err != nil {
		return s.dispatcher.NewInternalServerError("Failed to initiate OAuth", nil, "failed to get authorization URL", "provider", provider, "err", err)
	}

	return c.Redirect(http.StatusFound, authURL)
}

// OAuthCallback handles OAuth callback after user approves
func (s *AuthService) OAuthCallback(c echo.Context) error {
	provider := c.Param("provider")
	_, ok := s.OAuthProviders[provider]
	if !ok {
		return s.dispatcher.NewBadRequest("Invalid OAuth provider", nil)
	}

	// Add provider context to the request
	req := gothic.GetContextWithProvider(c.Request(), provider)
	c.SetRequest(req)

	// Get the gothic user from the callback using the request and response
	gothUser, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return s.dispatcher.NewUnauthorized("Failed to authorize with OAuth provider", nil)
	}

	// Check if user already exists
	existingUser, err := s.db.User.GetByEmailOrUsername(c.Request().Context(), user.GetByEmailOrUsernameParams{
		Email:    gothUser.Email,
		Username: generateUsernameFromEmail(gothUser.Email),
	})

	var userId int64
	if err == sql.ErrNoRows {
		// User doesn't exist, create a new one
		// Generate a random password for OAuth users
		randomPassword, err := uuid.NewV4()
		if err != nil {
			return s.dispatcher.NewInternalServerError("Failed to generate user", err)
		}

		bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(randomPassword.String()), bcrypt.DefaultCost)
		if err != nil {
			return s.dispatcher.NewInternalServerError("Failed to hash password", err)
		}

		newUser, err := s.db.User.UpsertUser(c.Request().Context(), user.UpsertUserParams{
			Username: generateUsernameFromEmail(gothUser.Email),
			Email:    gothUser.Email,
			Password: string(bcryptPassword),
			Role:     "user",
		})
		if err != nil {
			return s.dispatcher.NewInternalServerError("Failed to create user", err)
		}
		userId = newUser.ID
	} else if err != nil {
		return s.dispatcher.NewInternalServerError("Failed to check user", err)
	} else {
		// User exists, use their ID
		userId = existingUser.ID
	}

	// Generate session cookie
	if err := s.generateCookie(c, userId); err != nil {
		return s.dispatcher.NewInternalServerError("Failed to generate session", err)
	}

	// Redirect to dashboard on frontend
	return c.Redirect(http.StatusFound, models.ENV_VARS.FRONTEND_DASH)
}

// generateCookie generates a session cookie for the user
func (s *AuthService) generateCookie(c echo.Context, userId int64) error {
	uid, err := uuid.NewV4()
	if err != nil {
		return s.dispatcher.NewInternalServerError("Failed to generate session token", err)
	}

	_, err = s.db.Session.UpsertSession(c.Request().Context(), dbsessions.UpsertSessionParams{
		UserID:       userId,
		SessionToken: uid.String(),
	})
	if err != nil {
		return s.dispatcher.NewInternalServerError("Failed to create session", err)
	}

	cookie := http.Cookie{
		Name:     models.COOKIE_NAME,
		Value:    uid.String(),
		Path:     "/",
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   true, // Set to false for HTTP localhost
		SameSite: http.SameSiteNoneMode,
	}
	c.SetCookie(&cookie)

	return nil
}

// generateUsernameFromEmail generates a username from an email address
func generateUsernameFromEmail(email string) string {
	// Extract the part before @ and clean it
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		username := parts[0]
		// Replace dots and underscores with hyphens
		username = strings.ReplaceAll(username, ".", "-")
		return username
	}
	return "user-" + uuid.Must(uuid.NewV4()).String()[:8]
}
