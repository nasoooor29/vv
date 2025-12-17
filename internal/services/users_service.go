package services

import (
	"fmt"
	"log/slog"
	"net/http"

	"visory/internal/database"
	"visory/internal/database/user"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type UsersService struct {
	db     *database.Service
	logger *slog.Logger
}

// NewUsersService creates a new UsersService with dependency injection
func NewUsersService(db *database.Service, logger *slog.Logger) *UsersService {
	// Create a grouped logger for users service
	usersLogger := logger.WithGroup("users")
	return &UsersService{
		db:     db,
		logger: usersLogger,
	}
}

// GetAllUsers returns all users
func (s *UsersService) GetAllUsers(c echo.Context) error {
	users, err := s.db.User.GetAllUsers(c.Request().Context())
	if err != nil {
		s.logger.Error("error fetching users", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch users").SetInternal(err)
	}

	return c.JSON(http.StatusOK, users)
}

// GetUserById returns a user by ID
func (s *UsersService) GetUserById(c echo.Context) error {
	userID := c.Param("id")
	var id int64
	fmt.Sscanf(userID, "%d", &id)

	// Note: You may need to add a GetUserById method to your database layer
	// For now, this is a placeholder that assumes such a method exists
	users, err := s.db.User.GetAllUsers(c.Request().Context())
	if err != nil {
		s.logger.Error("error fetching users", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch users").SetInternal(err)
	}

	for _, u := range users {
		if u.ID == id {
			return c.JSON(http.StatusOK, u)
		}
	}

	return echo.NewHTTPError(http.StatusNotFound, "User not found")
}

// CreateUser creates a new user
func (s *UsersService) CreateUser(c echo.Context) error {
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
		s.logger.Error("user already exists", "email", p.Email, "username", p.Username)
		return echo.NewHTTPError(http.StatusConflict, "User already exists")
	}

	// If user doesn't have a role, assign "user" role
	if p.Role == "" {
		p.Role = "user"
	}

	// Hash password
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("error hashing password", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password").SetInternal(err)
	}
	p.Password = string(bcryptPassword)

	newUser, err := s.db.User.CreateUser(c.Request().Context(), p)
	if err != nil {
		s.logger.Error("error creating user", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user").SetInternal(err)
	}

	return c.JSON(http.StatusOK, newUser)
}

// UpdateUser updates an existing user
func (s *UsersService) UpdateUser(c echo.Context) error {
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
		s.logger.Error("error updating user", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user").SetInternal(err)
	}

	return c.JSON(http.StatusOK, updatedUser)
}

// DeleteUser deletes a user by ID
func (s *UsersService) DeleteUser(c echo.Context) error {
	userID := c.Param("id")
	var id int64
	fmt.Sscanf(userID, "%d", &id)

	err := s.db.User.DeleteUser(c.Request().Context(), id)
	if err != nil {
		s.logger.Error("error deleting user", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete user").SetInternal(err)
	}

	return c.NoContent(http.StatusOK)
}

// UpdateUserRole updates a user's role
func (s *UsersService) UpdateUserRole(c echo.Context) error {
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
	if err != nil {
		s.logger.Error("error updating user role", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user role").SetInternal(err)
	}

	return c.JSON(http.StatusOK, updatedUser)
}
