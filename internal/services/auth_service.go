package services

import (
	"database/sql"
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
	logger         *slog.Logger
	OAuthProviders map[string]goth.Provider
}

func (s *AuthService) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie(models.COOKIE_NAME)
		if err != nil {
			s.logger.Error("error happened", "err", err)
			return echo.NewHTTPError(http.StatusUnauthorized, "Failed to get user by session token").SetInternal(err)
		}
		_, err = s.db.User.GetBySessionToken(c.Request().Context(), cookie.Value)
		if err != nil {
			s.logger.Error("error happened", "err", err)
			return echo.NewHTTPError(http.StatusUnauthorized, "Failed to get user by session token").SetInternal(err)
		}

		return next(c)
	}
}

func (s *AuthService) RBACMiddleware(policies ...models.RBACPolicy) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(models.COOKIE_NAME)
			if err != nil {
				s.logger.Error("error happened", "err", err)
				return echo.NewHTTPError(http.StatusUnauthorized, "Failed to get user by session token").SetInternal(err)
			}
			user, err := s.db.User.GetBySessionToken(c.Request().Context(), cookie.Value)
			if err != nil {
				s.logger.Error("error happened", "err", err)
				return echo.NewHTTPError(http.StatusUnauthorized, "Failed to get user by session token").SetInternal(err)
			}

			user_roles := models.RoleToRBACPolicies(user.Role)
			if _, ok := user_roles[models.RBAC_USER_ADMIN]; ok {
				return next(c)
			}
			for _, policy := range policies {
				if v, ok := user_roles[policy]; !ok || !v {
					return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
				}
			}

			return next(c)
		}
	}
}

// NewAuthService creates a new AuthService with dependency injection
func NewAuthService(db *database.Service, logger *slog.Logger) *AuthService {
	// Create a grouped logger for auth service
	authLogger := logger.WithGroup("auth")
	authService := &AuthService{
		db:     db,
		logger: authLogger,
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
		s.logger.Error("error happened", "err", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Failed to get user by session token").SetInternal(err)
	}
	userWithSession, err := s.db.User.GetUserAndSessionByToken(c.Request().Context(), cookie.Value)
	if err != nil {
		s.logger.Error("error happened", "err", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Failed to get user by session token").SetInternal(err)
	}
	return c.JSON(http.StatusOK, userWithSession)
}

// Logout logs out the current user
func (s *AuthService) Logout(c echo.Context) error {
	cookie, err := c.Cookie(models.COOKIE_NAME)
	if err != nil {
		s.logger.Error("error happened", "err", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid cookie").SetInternal(err)
	}

	if err := s.db.Session.DeleteBySessionToken(c.Request().Context(), cookie.Value); err != nil {
		s.logger.Error("error happened", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to logout").SetInternal(err)
	}

	cookie.MaxAge = -1 // Expire the cookie
	c.SetCookie(cookie)

	return c.NoContent(http.StatusOK)
}

// Register registers a new user
func (s *AuthService) Register(c echo.Context) error {
	p := user.UpsertUserParams{}
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body").SetInternal(err)
	}
	_, err := s.db.User.GetByEmailOrUsername(c.Request().Context(), user.GetByEmailOrUsernameParams{
		Email:    p.Email,
		Username: p.Username,
	})
	if err == nil {
		s.logger.Error("user already exists", "email", p.Email, "username", p.Username)
		return echo.NewHTTPError(http.StatusConflict, "User already exists")
	}
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("error hashing password", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password").SetInternal(err)
	}
	p.Password = string(bcryptPassword)
	val, err := s.db.User.UpsertUser(c.Request().Context(), p)
	if err != nil {
		s.logger.Error("error happened", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user").SetInternal(err)
	}
	if err := s.generateCookie(c, val.ID); err != nil {
		s.logger.Error("error generating cookie", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate cookie").SetInternal(err)
	}

	return c.JSON(http.StatusOK, val)
}

// Login logs in a user
func (s *AuthService) Login(c echo.Context) error {
	p := models.Login{}
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body").SetInternal(err)
	}
	val, err := s.db.User.GetByEmailOrUsername(c.Request().Context(), user.GetByEmailOrUsernameParams{
		Email:    p.Username,
		Username: p.Username,
	})
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "You don't have an account please register").SetInternal(err)
	}
	if err != nil {
		s.logger.Error("error happened", "err", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Failed to login user").SetInternal(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(val.Password), []byte(p.Password))
	if err != nil {
		s.logger.Error("your username or password is wrong", "err", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "your username or password is wrong").SetInternal(err)
	}

	if err := s.generateCookie(c, val.ID); err != nil {
		s.logger.Error("error generating cookie", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate cookie").SetInternal(err)
	}

	return c.JSON(http.StatusOK, val)
}

// OAuthLogin handles OAuth login initiation
func (s *AuthService) OAuthLogin(c echo.Context) error {
	provider := c.Param("provider")
	_, ok := s.OAuthProviders[provider]
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid OAuth provider")
	}

	// Add provider context to the request
	req := gothic.GetContextWithProvider(c.Request(), provider)
	c.SetRequest(req)

	// Get the OAuth URL using gothic
	authURL, err := gothic.GetAuthURL(c.Response(), c.Request())
	if err != nil {
		s.logger.Error("failed to get authorization URL", "provider", provider, "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to initiate OAuth")
	}

	return c.Redirect(http.StatusFound, authURL)
}

// OAuthCallback handles OAuth callback after user approves
func (s *AuthService) OAuthCallback(c echo.Context) error {
	provider := c.Param("provider")
	_, ok := s.OAuthProviders[provider]
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid OAuth provider")
	}

	// Add provider context to the request
	req := gothic.GetContextWithProvider(c.Request(), provider)
	c.SetRequest(req)

	// Get the gothic user from the callback using the request and response
	gothUser, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		s.logger.Error("failed to complete user auth", "provider", provider, "err", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Failed to authorize with OAuth provider")
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
			s.logger.Error("failed to generate random password", "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate user")
		}

		bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(randomPassword.String()), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error("error hashing password", "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password")
		}

		newUser, err := s.db.User.UpsertUser(c.Request().Context(), user.UpsertUserParams{
			Username: generateUsernameFromEmail(gothUser.Email),
			Email:    gothUser.Email,
			Password: string(bcryptPassword),
			Role:     "user",
		})
		if err != nil {
			s.logger.Error("failed to create user", "email", gothUser.Email, "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
		}
		userId = newUser.ID
	} else if err != nil {
		s.logger.Error("failed to check user existence", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to check user")
	} else {
		// User exists, use their ID
		userId = existingUser.ID
	}

	// Generate session cookie
	if err := s.generateCookie(c, userId); err != nil {
		s.logger.Error("error generating cookie", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate session")
	}

	// Redirect to dashboard on frontend
	return c.Redirect(http.StatusFound, models.ENV_VARS.FRONTEND_DASH)
}

// generateCookie generates a session cookie for the user
func (s *AuthService) generateCookie(c echo.Context, userId int64) error {
	uid, err := uuid.NewV4()
	if err != nil {
		s.logger.Error("error happened", "err", err)
		return err
	}

	_, err = s.db.Session.UpsertSession(c.Request().Context(), dbsessions.UpsertSessionParams{
		UserID:       userId,
		SessionToken: uid.String(),
	})
	if err != nil {
		s.logger.Error("error happened", "err", err)
		return err
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
