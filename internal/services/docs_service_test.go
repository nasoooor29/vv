package services

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"visory/internal/utils"
)

func setupDocsServiceTest(t *testing.T) *DocsService {
	dispatcher := utils.NewDispatcher(nil)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	return &DocsService{
		db:         nil,
		Dispatcher: dispatcher,
		Logger:     logger,
	}
}

func TestServeSwagger(t *testing.T) {
	service := setupDocsServiceTest(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/docs/swagger", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := service.ServeSwagger(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, rec.Body.String(), "swagger-ui")
	assert.Contains(t, rec.Body.String(), "/api/docs/spec")
}

func TestServeSwaggerContainsSwaggerUI(t *testing.T) {
	service := setupDocsServiceTest(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/docs/swagger", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := service.ServeSwagger(c)
	assert.NoError(t, err)

	body := rec.Body.String()
	assert.Contains(t, body, "swagger-ui-bundle.js", "should include Swagger UI bundle")
	assert.Contains(t, body, "swagger-ui-standalone-preset.js", "should include Swagger UI preset")
	assert.Contains(t, body, "SwaggerUIBundle", "should initialize Swagger UI")
	assert.Contains(t, body, "Visory API - Swagger UI", "should have correct title")
}

func TestServeRedoc(t *testing.T) {
	service := setupDocsServiceTest(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/docs/redoc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := service.ServeRedoc(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, rec.Body.String(), "redoc")
}

func TestServeRedocContainsRedoc(t *testing.T) {
	service := setupDocsServiceTest(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/docs/redoc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := service.ServeRedoc(c)
	assert.NoError(t, err)

	body := rec.Body.String()
	assert.Contains(t, body, "redoc.standalone.js", "should include ReDoc standalone")
	assert.Contains(t, body, "<redoc", "should have redoc element")
	assert.Contains(t, body, "/api/docs/spec", "should reference spec URL")
	assert.Contains(t, body, "Visory API - ReDoc", "should have correct title")
}

func TestServeSpecReturnsValidJSON(t *testing.T) {
	service := setupDocsServiceTest(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/docs/spec", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := service.ServeSpec(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify the response is valid JSON
	var spec map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &spec)
	assert.NoError(t, err, "response should be valid JSON")
	assert.NotEmpty(t, spec)
}

func TestServeSpecContainsAPIInfo(t *testing.T) {
	service := setupDocsServiceTest(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/docs/spec", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := service.ServeSpec(c)
	require.NoError(t, err)

	var spec map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &spec)
	require.NoError(t, err)

	// Check basic Swagger/OpenAPI structure
	assert.NotNil(t, spec["info"], "should have info section")
	assert.NotNil(t, spec["paths"], "should have paths section")

	info := spec["info"].(map[string]interface{})
	assert.Contains(t, info, "title")
	assert.Contains(t, info, "version")
}

func TestServeSpecContainsEndpoints(t *testing.T) {
	service := setupDocsServiceTest(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/docs/spec", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := service.ServeSpec(c)
	require.NoError(t, err)

	var spec map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &spec)
	require.NoError(t, err)

	paths := spec["paths"].(map[string]interface{})

	// Check for essential endpoints
	expectedEndpoints := []string{
		"/",
		"/health",
		"/auth/login",
		"/auth/register",
		"/logs",
		"/metrics",
		"/storage/devices",
		"/users",
	}

	for _, endpoint := range expectedEndpoints {
		assert.Contains(t, paths, endpoint, "should have %s endpoint", endpoint)
	}
}

func TestServeSpecContainsDefinitions(t *testing.T) {
	service := setupDocsServiceTest(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/docs/spec", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := service.ServeSpec(c)
	require.NoError(t, err)

	var spec map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &spec)
	require.NoError(t, err)

	definitions := spec["definitions"].(map[string]interface{})
	assert.NotEmpty(t, definitions, "should have type definitions")

	// Check for essential model definitions
	assert.Contains(t, definitions, "visory_internal_models.HTTPError")
}

func TestServeSpecJSONStructure(t *testing.T) {
	service := setupDocsServiceTest(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/docs/spec", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := service.ServeSpec(c)
	require.NoError(t, err)

	var spec map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &spec)
	require.NoError(t, err)

	// Verify OpenAPI/Swagger structure
	assert.Contains(t, spec, "swagger", "should have swagger field")
	assert.Contains(t, spec, "info", "should have info field")
	assert.Contains(t, spec, "paths", "should have paths field")
	assert.Contains(t, spec, "definitions", "should have definitions field")

	// Verify info details
	info := spec["info"].(map[string]interface{})
	title, ok := info["title"].(string)
	assert.True(t, ok && len(title) > 0, "info.title should be a non-empty string")

	version, ok := info["version"].(string)
	assert.True(t, ok && len(version) > 0, "info.version should be a non-empty string")
}

func TestServeSpecWithMultipleRequests(t *testing.T) {
	service := setupDocsServiceTest(t)
	e := echo.New()

	// Make multiple requests to ensure consistency
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/docs/spec", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := service.ServeSpec(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var spec map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &spec)
		assert.NoError(t, err)
		assert.NotEmpty(t, spec)
	}
}

func TestServeSpecContentTypeHeader(t *testing.T) {
	service := setupDocsServiceTest(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/docs/spec", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := service.ServeSpec(c)
	assert.NoError(t, err)

	contentType := rec.Header().Get("Content-Type")
	assert.Equal(t, "application/json", contentType, "Content-Type should be application/json")
}

func TestDocsServiceInitialization(t *testing.T) {
	service := setupDocsServiceTest(t)

	assert.NotNil(t, service)
	assert.NotNil(t, service.Dispatcher)
	assert.NotNil(t, service.Logger)
}
