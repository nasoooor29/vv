package utils

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestNoInputNoReturn(t *testing.T) {
	api := New()
	
	dummyHandler := func(c echo.Context) error { return nil }
	
	// Test 1: POST with no input - should warn about missing input
	api.add(http.MethodPost, "/test/missing-input", dummyHandler).
		Returns(http.StatusOK, nil)
	
	// Test 2: Route with no responses - should warn about missing response
	api.add(http.MethodGet, "/test/missing-response", dummyHandler)
	
	// Test 3: Explicit NoInput - should not warn about input
	api.add(http.MethodGet, "/test/no-input-explicit", dummyHandler).
		NoInput().
		Returns(http.StatusOK, nil)
	
	// Test 4: Explicit NoReturn - should not warn about response
	api.add(http.MethodPost, "/test/no-return-explicit", dummyHandler).
		Input(map[string]interface{}{}).
		NoReturn()
	
	// Generate OpenAPI spec (should log warnings for tests 1 and 2)
	spec := api.OpenAPI()
	
	if spec == nil {
		t.Fatal("OpenAPI spec should not be nil")
	}
	
	if len(api.routes) != 4 {
		t.Fatalf("Expected 4 routes, got %d", len(api.routes))
	}
	
	// Verify the flags are set correctly
	if !api.routes[2].noInput {
		t.Error("Route 2 should have noInput flag set")
	}
	
	if !api.routes[3].noReturn {
		t.Error("Route 3 should have noReturn flag set")
	}
	
	t.Log("All tests passed!")
}
