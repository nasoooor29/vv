package server

import (
	"database/sql"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strings"

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

// InitializeOAuth sets up OAuth providers with environment variables
func InitializeOAuth() map[string]goth.Provider {
	// env := models.ENV_VARS
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

// OAuthLogin handles OAuth login initiation
func (s *Server) OAuthLogin(c echo.Context) error {
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
		slog.Error("failed to get authorization URL", "provider", provider, "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to initiate OAuth")
	}

	return c.Redirect(http.StatusFound, authURL)
}

// OAuthCallback handles OAuth callback after user approves
func (s *Server) OAuthCallback(c echo.Context) error {
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
		slog.Error("failed to complete user auth", "provider", provider, "err", err)
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
			slog.Error("failed to generate random password", "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate user")
		}

		bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(randomPassword.String()), bcrypt.DefaultCost)
		if err != nil {
			slog.Error("error hashing password", "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password")
		}

		newUser, err := s.db.User.UpsertUser(c.Request().Context(), user.UpsertUserParams{
			Username: generateUsernameFromEmail(gothUser.Email),
			Email:    gothUser.Email,
			Password: string(bcryptPassword),
			Role:     "user",
		})
		if err != nil {
			slog.Error("failed to create user", "email", gothUser.Email, "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
		}
		userId = newUser.ID
	} else if err != nil {
		slog.Error("failed to check user existence", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to check user")
	} else {
		// User exists, use their ID
		userId = existingUser.ID
	}

	// Generate session cookie
	if _, err := s.GenerateCookie(c, userId); err != nil {
		slog.Error("error generating cookie", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate session")
	}

	// Redirect to dashboard on frontend
	return c.Redirect(http.StatusFound, models.ENV_VARS.FRONTEND_DASH)
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
