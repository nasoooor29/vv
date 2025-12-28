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
	"visory/internal/database/user"
	"visory/internal/models"
	"visory/internal/utils"

	dbsessions "visory/internal/database/sessions"

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
	Dispatcher     *utils.Dispatcher
	Logger         *slog.Logger
	OAuthProviders map[string]goth.Provider
}

func (s *AuthService) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie(models.COOKIE_NAME)
		if err != nil {
			return s.Dispatcher.NewUnauthorized("Failed to get user by session token", err)
		}
		userWithSession, err := s.db.User.GetUserAndSessionByToken(c.Request().Context(), cookie.Value)
		if err != nil {
			return s.Dispatcher.NewUnauthorized("Failed to get user by session token", err)
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
				return s.Dispatcher.NewUnauthorized("Failed to get user by session token", err)
			}
			user, err := s.db.User.GetBySessionToken(c.Request().Context(), cookie.Value)
			if err != nil {
				return s.Dispatcher.NewUnauthorized("Failed to get user by session token", err)
			}

			user_roles := models.RoleToRBACPolicies(user.Role)
			if _, ok := user_roles[models.RBAC_USER_ADMIN]; ok {
				return next(c)
			}
			for _, policy := range policies {
				if v, ok := user_roles[policy]; !ok || !v {
					return s.Dispatcher.NewForbidden("Insufficient permissions", err)
				}
			}

			return next(c)
		}
	}
}

// NewAuthService creates a new AuthService with dependency injection
func NewAuthService(db *database.Service, dispatcher *utils.Dispatcher, logger *slog.Logger) *AuthService {
	// Create a grouped logger for auth service
	authService := &AuthService{
		db:         db,
		Dispatcher: dispatcher.WithGroup("auth"),
		Logger:     logger.WithGroup("auth"),
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

//	@Summary      my info
//	@Description  Get current user info
//	@Tags         accounts
//	@Produce      json
//	@Success      200  {object}  user.GetUserAndSessionByTokenRow
//	@Failure      401  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /auth/me [get]
//
// Me returns the current authenticated user
func (s *AuthService) Me(c echo.Context) error {
	cookie, err := c.Cookie(models.COOKIE_NAME)
	if err != nil {
		return s.Dispatcher.NewUnauthorized("Failed to get user by session token", err)
	}
	userWithSession, err := s.db.User.GetUserAndSessionByToken(c.Request().Context(), cookie.Value)
	if err != nil {
		return s.Dispatcher.NewUnauthorized("Failed to get user by session token", err)
	}
	return c.JSON(http.StatusOK, userWithSession)
}

//	@Summary      logout
//	@Description  delete the session and logout the current user
//	@Tags         accounts
//	@Produce      json
//	@Success      200  {null} null
//	@Failure      401  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /auth/logout [get]
//
// Logout logs out the current user
func (s *AuthService) Logout(c echo.Context) error {
	cookie, err := c.Cookie(models.COOKIE_NAME)
	if err != nil {
		return s.Dispatcher.NewBadRequest("Invalid cookie", err)
	}

	if err := s.db.Session.DeleteBySessionToken(c.Request().Context(), cookie.Value); err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to logout", err)
	}

	cookie.MaxAge = -1 // Expire the cookie
	c.SetCookie(cookie)

	return c.NoContent(http.StatusOK)
}

// @Summary      register
// @Description  registers a new user
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        user  body      user.UpsertUserParams  true  "User registration info"
// @Success      200   {object}  user.GetUserAndSessionByTokenRow
// @Failure      400   {object}  models.HTTPError
// @Failure      409   {object}  models.HTTPError
// @Failure      500   {object}  models.HTTPError
// @Router       /auth/register [post]
// Register registers a new user
func (s *AuthService) Register(c echo.Context) error {
	p := user.UpsertUserParams{}
	if err := c.Bind(&p); err != nil {
		return s.Dispatcher.NewBadRequest("Invalid request body", err)
	}
	_, err := s.db.User.GetByEmailOrUsername(c.Request().Context(), user.GetByEmailOrUsernameParams{
		Email:    p.Email,
		Username: p.Username,
	})
	if err == nil {
		return s.Dispatcher.NewConflict("User already exists", nil)
	}
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to hash password", err)
	}
	p.Password = string(bcryptPassword)

	// Check if this is the first user - if so, make them admin
	userCount, err := s.db.User.CountUsers(c.Request().Context())
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to check user count", err)
	}
	if userCount == 0 {
		p.Role = string(models.RBAC_USER_ADMIN)
		s.Logger.Info("First user registered, granting admin role", "email", p.Email)
	}

	val, err := s.db.User.UpsertUser(c.Request().Context(), p)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to register user", err)
	}
	sessionToken, err := s.generateCookie(c, val.ID)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to generate cookie", err)
	}

	// Fetch the full user with session to return consistent response
	userWithSession, err := s.db.User.GetUserAndSessionByToken(c.Request().Context(), sessionToken)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to get user session", err)
	}

	return c.JSON(http.StatusOK, userWithSession)
}

// @Summary      login
// @Description  login with username/email and password
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        credentials  body      models.Login  true  "User login credentials"
// @Success      200   {object}  user.GetUserAndSessionByTokenRow
// @Failure      400   {object}  models.HTTPError
// @Failure      401   {object}  models.HTTPError
// @Failure      404   {object}  models.HTTPError
// @Failure      500   {object}  models.HTTPError
// @Router       /auth/login [post]
// Login logs in a user
func (s *AuthService) Login(c echo.Context) error {
	p := models.Login{}
	if err := c.Bind(&p); err != nil {
		return s.Dispatcher.NewBadRequest("Invalid request body", err)
	}
	val, err := s.db.User.GetByEmailOrUsername(c.Request().Context(), user.GetByEmailOrUsernameParams{
		Email:    p.Username,
		Username: p.Username,
	})
	if err == sql.ErrNoRows {
		return s.Dispatcher.NewNotFound("You don't have an account please register", err)
	}
	if err != nil {
		return s.Dispatcher.NewUnauthorized("Failed to login user", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(val.Password), []byte(p.Password))
	if err != nil {
		return s.Dispatcher.NewUnauthorized("your username or password is wrong", err)
	}

	sessionToken, err := s.generateCookie(c, val.ID)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to generate cookie", err)
	}

	// Fetch the full user with session to return consistent response
	userWithSession, err := s.db.User.GetUserAndSessionByToken(c.Request().Context(), sessionToken)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to get user session", err)
	}

	return c.JSON(http.StatusOK, userWithSession)
}

// @Summary      OAuth login
// @Description  initiate OAuth login flow
// @Tags         accounts
// @Produce      json
// @Param        provider  path      string  true  "OAuth provider (google, github)"
// @Failure      400   {object}  models.HTTPError
// @Failure      500   {object}  models.HTTPError
// @Router       /auth/oauth/{provider} [get]
// OAuthLogin handles OAuth login initiation
func (s *AuthService) OAuthLogin(c echo.Context) error {
	provider := c.Param("provider")
	_, ok := s.OAuthProviders[provider]
	if !ok {
		return s.Dispatcher.NewBadRequest("Invalid OAuth provider", fmt.Errorf("provider %s not supported", provider))
	}

	// Add provider context to the request
	req := gothic.GetContextWithProvider(c.Request(), provider)
	c.SetRequest(req)

	// Get the OAuth URL using gothic
	authURL, err := gothic.GetAuthURL(c.Response(), c.Request())
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to initiate OAuth", nil, "failed to get authorization URL", "provider", provider, "err", err)
	}

	return c.Redirect(http.StatusFound, authURL)
}

// OAuthCallback handles OAuth callback after user approves
func (s *AuthService) OAuthCallback(c echo.Context) error {
	provider := c.Param("provider")
	_, ok := s.OAuthProviders[provider]
	if !ok {
		return s.Dispatcher.NewBadRequest("Invalid OAuth provider", nil)
	}

	// Add provider context to the request
	req := gothic.GetContextWithProvider(c.Request(), provider)
	c.SetRequest(req)

	// Get the gothic user from the callback using the request and response
	gothUser, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return s.Dispatcher.NewUnauthorized("Failed to authorize with OAuth provider", nil)
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
			return s.Dispatcher.NewInternalServerError("Failed to generate user", err)
		}

		bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(randomPassword.String()), bcrypt.DefaultCost)
		if err != nil {
			return s.Dispatcher.NewInternalServerError("Failed to hash password", err)
		}

		// Check if this is the first user - if so, make them admin
		role := "user"
		userCount, err := s.db.User.CountUsers(c.Request().Context())
		if err != nil {
			return s.Dispatcher.NewInternalServerError("Failed to check user count", err)
		}
		if userCount == 0 {
			role = string(models.RBAC_USER_ADMIN)
			s.Logger.Info("First user via OAuth, granting admin role", "email", gothUser.Email)
		}

		newUser, err := s.db.User.UpsertUser(c.Request().Context(), user.UpsertUserParams{
			Username: generateUsernameFromEmail(gothUser.Email),
			Email:    gothUser.Email,
			Password: string(bcryptPassword),
			Role:     role,
		})
		if err != nil {
			return s.Dispatcher.NewInternalServerError("Failed to create user", err)
		}
		userId = newUser.ID
	} else if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to check user", err)
	} else {
		// User exists, use their ID
		userId = existingUser.ID
	}

	// Generate session cookie
	if _, err := s.generateCookie(c, userId); err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to generate session", err)
	}

	// Redirect to dashboard on frontend
	return c.Redirect(http.StatusFound, models.ENV_VARS.FRONTEND_DASH)
}

// generateCookie generates a session cookie for the user and returns the session token
func (s *AuthService) generateCookie(c echo.Context, userId int64) (string, error) {
	uid, err := uuid.NewV4()
	if err != nil {
		return "", s.Dispatcher.NewInternalServerError("Failed to generate session token", err)
	}

	_, err = s.db.Session.UpsertSession(c.Request().Context(), dbsessions.UpsertSessionParams{
		UserID:       userId,
		SessionToken: uid.String(),
	})
	if err != nil {
		return "", s.Dispatcher.NewInternalServerError("Failed to create session", err)
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

	return uid.String(), nil
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
