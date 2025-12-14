package server

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"visory/internal/database/sessions"
	"visory/internal/database/user"
	"visory/internal/models"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) Me(c echo.Context) error {
	cookie, err := c.Cookie(models.COOKIE_NAME)
	if err != nil {
		slog.Error("error happened", "err", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Failed to get user by session token").SetInternal(err)
	}
	userWithSession, err := s.db.User.GetUserAndSessionByToken(c.Request().Context(), cookie.Value)
	if err != nil {
		slog.Error("error happened", "err", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Failed to get user by session token").SetInternal(err)
	}
	return c.JSON(http.StatusOK, userWithSession)
}

func (s *Server) Logout(c echo.Context) error {
	cookie, err := c.Cookie(models.COOKIE_NAME)
	if err != nil {
		slog.Error("error happened", "err", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid cookie").SetInternal(err)
	}

	if err := s.db.Session.DeleteBySessionToken(c.Request().Context(), cookie.Value); err != nil {
		slog.Error("error happened", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to logout").SetInternal(err)
	}

	cookie.MaxAge = -1 // Expire the cookie
	c.SetCookie(cookie)

	return c.NoContent(http.StatusOK)
}

func (s *Server) Register(c echo.Context) error {
	p := user.UpsertUserParams{}
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body").SetInternal(err)
	}
	_, err := s.db.User.GetByEmailOrUsername(c.Request().Context(), user.GetByEmailOrUsernameParams{
		Email:    p.Email,
		Username: p.Username,
	})
	if err == nil {
		slog.Error("user already exists", "email", p.Email, "username", p.Username)
		return echo.NewHTTPError(http.StatusConflict, "User already exists")
	}
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("error hashing password", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password").SetInternal(err)
	}
	p.Password = string(bcryptPassword)
	val, err := s.db.User.UpsertUser(c.Request().Context(), p)
	if err != nil {
		slog.Error("error happened", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user").SetInternal(err)
	}
	if err := s.GenerateCookie(c, val.ID); err != nil {
		slog.Error("error generating cookie", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate cookie").SetInternal(err)
	}

	return c.JSON(http.StatusOK, val)
}

func (s *Server) GenerateCookie(c echo.Context, userId int64) error {
	uid, err := uuid.NewV4()
	if err != nil {
		slog.Error("error happened", "err", err)
		return err
	}

	_, err = s.db.Session.UpsertSession(c.Request().Context(), sessions.UpsertSessionParams{
		UserID:       userId,
		SessionToken: uid.String(),
	})
	if err != nil {
		slog.Error("error happened", "err", err)
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

func (s *Server) Login(c echo.Context) error {
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
		slog.Error("error happened", "err", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Failed to login user").SetInternal(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(val.Password), []byte(p.Password))
	if err != nil {
		slog.Error("your username or password is wrong", "err", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "your username or password is wrong").SetInternal(err)
	}

	if err := s.GenerateCookie(c, val.ID); err != nil {
		slog.Error("error generating cookie", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate cookie").SetInternal(err)
	}

	return c.JSON(http.StatusOK, val)
}

func (s *Server) GetAllUsers(c echo.Context) error {
	users, err := s.db.User.GetAllUsers(c.Request().Context())
	if err != nil {
		slog.Error("error fetching users", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch users").SetInternal(err)
	}

	return c.JSON(http.StatusOK, users)
}

func (s *Server) CreateUser(c echo.Context) error {
	p := user.CreateUserParams{}
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body").SetInternal(err)
	}

	// Check if user already exists
	_, err := s.db.User.GetByEmailOrUsername(c.Request().Context(), user.GetByEmailOrUsernameParams{
		Email:    p.Email,
		Username: p.Username,
	})
	if err == nil {
		slog.Error("user already exists", "email", p.Email, "username", p.Username)
		return echo.NewHTTPError(http.StatusConflict, "User already exists")
	}

	// If user doesn't have a role, assign "user" role
	if p.Role == "" {
		p.Role = "user"
	}

	// Hash password
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("error hashing password", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password").SetInternal(err)
	}
	p.Password = string(bcryptPassword)

	newUser, err := s.db.User.CreateUser(c.Request().Context(), p)
	if err != nil {
		slog.Error("error creating user", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user").SetInternal(err)
	}

	return c.JSON(http.StatusOK, newUser)
}

func (s *Server) UpdateUser(c echo.Context) error {
	userID := c.Param("id")
	p := struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}{}

	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body").SetInternal(err)
	}

	// Parse ID from URL param
	var id int64
	fmt.Sscanf(userID, "%d", &id)

	// If role is empty, assign "user" role
	if p.Role == "" {
		p.Role = "user"
	}

	updatedUser, err := s.db.User.UpdateUser(c.Request().Context(), user.UpdateUserParams{
		Username: p.Username,
		Email:    p.Email,
		Role:     p.Role,
		ID:       id,
	})
	if err != nil {
		slog.Error("error updating user", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user").SetInternal(err)
	}

	return c.JSON(http.StatusOK, updatedUser)
}

func (s *Server) DeleteUser(c echo.Context) error {
	userID := c.Param("id")
	var id int64
	fmt.Sscanf(userID, "%d", &id)

	err := s.db.User.DeleteUser(c.Request().Context(), id)
	if err != nil {
		slog.Error("error deleting user", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete user").SetInternal(err)
	}

	return c.NoContent(http.StatusOK)
}

func (s *Server) UpdateUserRole(c echo.Context) error {
	userID := c.Param("id")
	p := struct {
		Role string `json:"role"`
	}{}

	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body").SetInternal(err)
	}

	var id int64
	fmt.Sscanf(userID, "%d", &id)

	// If role is empty, assign "user" role
	if p.Role == "" {
		p.Role = "user"
	}

	updatedUser, err := s.db.User.UpdateUserRole(c.Request().Context(), user.UpdateUserRoleParams{
		Role: p.Role,
		ID:   id,
	})
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "User not found").SetInternal(err)
	}

	if err != nil {
		slog.Error("error updating user role", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user role").SetInternal(err)
	}

	return c.JSON(http.StatusOK, updatedUser)
}
