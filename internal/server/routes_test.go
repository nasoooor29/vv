package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"visory/internal/database"
	"visory/internal/services"
	"visory/internal/utils"

	"github.com/labstack/echo/v4"
)

func TestHandler(t *testing.T) {
	// Create mock dependencies
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create a server with minimal dependencies for testing
	db := database.New()
	mySlog := utils.NewMySlog(logger, db)
	s := &Server{
		port:           9999,
		logger:         mySlog,
		db:             db,
		authService:    services.NewAuthService(db, mySlog),
		usersService:   services.NewUsersService(db, logger),
		storageService: services.NewStorageService(logger),
		logsService:    services.NewLogsService(db, logger),
		metricsService: services.NewMetricsService(db, logger),
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()
	c := e.NewContext(req, resp)

	// Assertions
	if err := s.HelloWorldHandler(c); err != nil {
		t.Errorf("handler() error = %v", err)
		return
	}
	if resp.Code != http.StatusOK {
		t.Errorf("handler() wrong status code = %v", resp.Code)
		return
	}
	expected := map[string]string{"message": "Hello World"}
	var actual map[string]string
	// Decode the response body into the actual map
	if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
		t.Errorf("handler() error decoding response body: %v", err)
		return
	}
	// Compare the decoded response with the expected value
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("handler() wrong response body. expected = %v, actual = %v", expected, actual)
		return
	}
}
