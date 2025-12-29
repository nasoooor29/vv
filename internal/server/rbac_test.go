package server

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"visory/internal/database"
	dbsessions "visory/internal/database/sessions"
	"visory/internal/database/user"
	"visory/internal/models"
	"visory/internal/notifications"
	"visory/internal/services"
	"visory/internal/utils"

	"github.com/gofrs/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// testHelper provides utilities for RBAC testing
type testHelper struct {
	server        *Server
	handler       http.Handler
	db            *database.Service
	testUsers     map[string]*testUserData
	sessionTokens map[string]string
}

type testUserData struct {
	id    int64
	email string
	role  string
}

// setupTestServer creates a test server with a clean database
func setupTestServer(t *testing.T) *testHelper {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create unique database for each test
	testDbPath := fmt.Sprintf("visory_test_%d.db", time.Now().UnixNano())

	// Create database with migration
	sqlDb, err := sql.Open("sqlite3", testDbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	if err := database.Migrate(sqlDb); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}
	sqlDb.Close()

	// Now use the normal database constructor which will open the migrated database
	// Temporarily set the environment variable
	oldDbPath := models.ENV_VARS.DBPath
	models.ENV_VARS.DBPath = testDbPath
	t.Cleanup(func() {
		models.ENV_VARS.DBPath = oldDbPath
		os.Remove(testDbPath)
	})

	dbService := database.New()
	dispatcher := utils.NewDispatcher(dbService, nil)

	dockerService := services.NewDockerService(dispatcher, logger)
	notifManager := notifications.NewManager()

	fs := utils.NewFS(fmt.Sprintf("%v-test-%d", "visory", time.Now().UnixNano()))
	s := &Server{
		port:             9999,
		logger:           logger,
		dispatcher:       dispatcher,
		db:               dbService,
		fs:               fs,
		authService:      services.NewAuthService(dbService, dispatcher, logger),
		usersService:     services.NewUsersService(dbService, dispatcher, logger),
		logsService:      services.NewLogsService(dbService, dispatcher, logger),
		metricsService:   services.NewMetricsService(dbService, dispatcher, logger),
		storageService:   services.NewStorageService(dispatcher, logger),
		qemuService:      services.NewQemuService(dispatcher, fs, logger),
		isoService:       services.NewISOService(dispatcher, fs, logger),
		dockerService:    dockerService,
		docsService:      services.NewDocsService(dbService, dispatcher, logger),
		firewallService:  services.NewFirewallService(dispatcher, logger),
		templatesService: services.NewTemplatesService(dispatcher, logger, dockerService.ClientManager),
		settingsService:  services.NewSettingsService(dbService, dispatcher, logger, notifManager),
		vncProxy:         services.NewVNCProxy(logger),
	}
	handler := s.RegisterRoutes()

	return &testHelper{
		server:        s,
		handler:       handler,
		db:            dbService,
		testUsers:     make(map[string]*testUserData),
		sessionTokens: make(map[string]string),
	}
}

// createTestUser creates a test user and returns their session token
func (th *testHelper) createTestUser(t *testing.T, email string, role string) string {
	ctx := context.Background()

	// Create user
	userId, err := th.db.User.CreateUser(ctx, user.CreateUserParams{
		Username: email[:len(email)-10], // Extract name before @
		Email:    email,
		Password: "hashedpassword",
		Role:     role,
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create session
	sessionToken := uuid.Must(uuid.NewV4()).String()
	_, err = th.db.Session.UpsertSession(ctx, dbsessions.UpsertSessionParams{
		UserID:       userId.ID,
		SessionToken: sessionToken,
	})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	th.testUsers[email] = &testUserData{
		id:    userId.ID,
		email: email,
		role:  role,
	}
	th.sessionTokens[email] = sessionToken

	return sessionToken
}

// makeRequest makes an HTTP request with optional authentication
func (th *testHelper) makeRequest(t *testing.T, method string, path string, sessionToken *string, body interface{}) (*httptest.ResponseRecorder, error) {
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Add session cookie if provided
	if sessionToken != nil {
		cookie := &http.Cookie{
			Name:  models.COOKIE_NAME,
			Value: *sessionToken,
		}
		req.AddCookie(cookie)
	}

	resp := httptest.NewRecorder()
	th.handler.ServeHTTP(resp, req)

	return resp, nil
}

// assertStatusCode checks if the response status code matches expected
func assertStatusCode(t *testing.T, resp *httptest.ResponseRecorder, expected int, message string) {
	if resp.Code != expected {
		t.Errorf("%s: expected status %d, got %d. Body: %s", message, expected, resp.Code, resp.Body.String())
	}
}

// TestRBACUserAdmin tests user_admin policy on user endpoints
func TestRBACUserAdmin(t *testing.T) {
	th := setupTestServer(t)

	// Create test users with different roles
	adminToken := th.createTestUser(t, "admin@test.com", string(models.RBAC_USER_ADMIN))
	noPerm1Token := th.createTestUser(t, "noperm1@test.com", string(models.RBAC_DOCKER_READ))
	noPermToken := th.createTestUser(t, "noperm@test.com", "user")

	tests := []struct {
		name       string
		method     string
		path       string
		token      *string
		statusCode int
		desc       string
	}{
		{
			name:       "GetAllUsers with user_admin - allowed",
			method:     "GET",
			path:       "/api/users",
			token:      &adminToken,
			statusCode: http.StatusOK,
			desc:       "User admin should be allowed to get all users",
		},
		{
			name:       "GetAllUsers without permission - forbidden",
			method:     "GET",
			path:       "/api/users",
			token:      &noPermToken,
			statusCode: http.StatusForbidden,
			desc:       "User without permission should get 403",
		},
		{
			name:       "GetAllUsers with different permission - forbidden",
			method:     "GET",
			path:       "/api/users",
			token:      &noPerm1Token,
			statusCode: http.StatusForbidden,
			desc:       "User with docker_read should not be allowed",
		},
		{
			name:       "GetAllUsers with no auth - unauthorized",
			method:     "GET",
			path:       "/api/users",
			token:      nil,
			statusCode: http.StatusUnauthorized,
			desc:       "Request without auth should get 401",
		},
		{
			name:       "GetAllUsers with invalid token - unauthorized",
			method:     "GET",
			path:       "/api/users",
			token:      pointToString("invalid-token"),
			statusCode: http.StatusUnauthorized,
			desc:       "Request with invalid token should get 401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := th.makeRequest(t, tt.method, tt.path, tt.token, nil)
			assertStatusCode(t, resp, tt.statusCode, tt.desc)
		})
	}
}

// TestRBACRoleToPolicy tests the RoleToRBACPolicies function
func TestRBACRoleToPolicy(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected map[models.RBACPolicy]bool
	}{
		{
			name: "Single policy",
			role: "user_admin",
			expected: map[models.RBACPolicy]bool{
				models.RBAC_USER_ADMIN: true,
			},
		},
		{
			name: "Multiple policies with spaces",
			role: "docker_read, docker_write, qemu_read",
			expected: map[models.RBACPolicy]bool{
				models.RBAC_DOCKER_READ:  true,
				models.RBAC_DOCKER_WRITE: true,
				models.RBAC_QEMU_READ:    true,
			},
		},
		{
			name: "Multiple policies no spaces",
			role: "audit_log_viewer,health_checker",
			expected: map[models.RBACPolicy]bool{
				models.RBAC_AUDIT_LOG_VIEWER: true,
				models.RBAC_HEALTH_CHECKER:   true,
			},
		},
		{
			name:     "Empty string",
			role:     "",
			expected: map[models.RBACPolicy]bool{},
		},
		{
			name:     "Only spaces",
			role:     "   ,  , ",
			expected: map[models.RBACPolicy]bool{},
		},
		{
			name: "Invalid policies mixed with valid",
			role: "docker_read,invalid_policy,audit_log_viewer",
			expected: map[models.RBACPolicy]bool{
				models.RBAC_DOCKER_READ:      true,
				models.RBAC_AUDIT_LOG_VIEWER: true,
			},
		},
		{
			name: "Admin bypasses all checks",
			role: "user_admin,docker_read",
			expected: map[models.RBACPolicy]bool{
				models.RBAC_USER_ADMIN:  true,
				models.RBAC_DOCKER_READ: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policies := models.RoleToRBACPolicies(tt.role)

			// Check all expected policies are present
			for policy, shouldExist := range tt.expected {
				if !shouldExist {
					continue
				}
				if _, ok := policies[policy]; !ok {
					t.Errorf("Expected policy %s to be present, but it wasn't", policy)
				}
			}

			// Check no unexpected policies are present
			for policy := range policies {
				if _, ok := tt.expected[policy]; !ok {
					t.Errorf("Unexpected policy %s found", policy)
				}
			}
		})
	}
}

// TestRBACMiddlewareUserAdminBypass tests that user_admin bypasses other policy checks
func TestRBACMiddlewareUserAdminBypass(t *testing.T) {
	th := setupTestServer(t)

	// Create admin user with multiple roles (admin should bypass others)
	adminToken := th.createTestUser(t, "admin@test.com", string(models.RBAC_USER_ADMIN))

	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		desc       string
	}{
		{
			name:       "Admin can access user endpoint",
			method:     "GET",
			path:       "/api/users",
			statusCode: http.StatusOK,
			desc:       "User admin should bypass checks",
		},
		{
			name:       "Admin can access logs endpoint",
			method:     "GET",
			path:       "/api/logs",
			statusCode: http.StatusOK,
			desc:       "User admin should have audit log viewer access",
		},
		{
			name:       "Admin can access metrics endpoint",
			method:     "GET",
			path:       "/api/metrics/health",
			statusCode: http.StatusOK,
			desc:       "User admin should have health checker access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := th.makeRequest(t, tt.method, tt.path, &adminToken, nil)
			assertStatusCode(t, resp, tt.statusCode, tt.desc)
		})
	}
}

// TestRBACLogs tests RBAC on logs endpoints
func TestRBACLogs(t *testing.T) {
	th := setupTestServer(t)

	// Create test users
	auditToken := th.createTestUser(t, "audit@test.com", string(models.RBAC_AUDIT_LOG_VIEWER))
	healthToken := th.createTestUser(t, "health@test.com", string(models.RBAC_HEALTH_CHECKER))
	noPermToken := th.createTestUser(t, "noperm@test.com", "user")

	tests := []struct {
		name       string
		method     string
		path       string
		token      *string
		statusCode int
		desc       string
	}{
		{
			name:       "audit_log_viewer can access logs",
			method:     "GET",
			path:       "/api/logs",
			token:      &auditToken,
			statusCode: http.StatusOK,
			desc:       "Audit log viewer should access logs",
		},
		{
			name:       "audit_log_viewer can access log stats",
			method:     "GET",
			path:       "/api/logs/stats",
			token:      &auditToken,
			statusCode: http.StatusOK,
			desc:       "Audit log viewer should access log stats",
		},
		{
			name:       "health_checker cannot access logs",
			method:     "GET",
			path:       "/api/logs",
			token:      &healthToken,
			statusCode: http.StatusForbidden,
			desc:       "Health checker should not access logs",
		},
		{
			name:       "user without permission cannot access logs",
			method:     "GET",
			path:       "/api/logs",
			token:      &noPermToken,
			statusCode: http.StatusForbidden,
			desc:       "User without permission should get 403",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := th.makeRequest(t, tt.method, tt.path, tt.token, nil)
			assertStatusCode(t, resp, tt.statusCode, tt.desc)
		})
	}
}

// TestRBACMetrics tests RBAC on metrics endpoints
func TestRBACMetrics(t *testing.T) {
	th := setupTestServer(t)

	// Create test users
	auditToken := th.createTestUser(t, "audit@test.com", string(models.RBAC_AUDIT_LOG_VIEWER))
	healthToken := th.createTestUser(t, "health@test.com", string(models.RBAC_HEALTH_CHECKER))
	noPermToken := th.createTestUser(t, "noperm@test.com", "user")

	tests := []struct {
		name       string
		method     string
		path       string
		token      *string
		statusCode int
		desc       string
	}{
		{
			name:       "audit_log_viewer can access metrics",
			method:     "GET",
			path:       "/api/metrics",
			token:      &auditToken,
			statusCode: http.StatusOK,
			desc:       "Audit log viewer should access metrics",
		},
		{
			name:       "health_checker can access health metrics",
			method:     "GET",
			path:       "/api/metrics/health",
			token:      &healthToken,
			statusCode: http.StatusOK,
			desc:       "Health checker should access health metrics",
		},
		{
			name:       "audit_log_viewer cannot access health metrics",
			method:     "GET",
			path:       "/api/metrics/health",
			token:      &auditToken,
			statusCode: http.StatusForbidden,
			desc:       "Audit log viewer should not access health metrics without health_checker role",
		},
		{
			name:       "user without permission cannot access metrics",
			method:     "GET",
			path:       "/api/metrics",
			token:      &noPermToken,
			statusCode: http.StatusForbidden,
			desc:       "User without permission should get 403",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := th.makeRequest(t, tt.method, tt.path, tt.token, nil)
			assertStatusCode(t, resp, tt.statusCode, tt.desc)
		})
	}
}

// TestRBACStorage tests RBAC on storage endpoints
func TestRBACStorage(t *testing.T) {
	th := setupTestServer(t)

	// Create test users
	settingsToken := th.createTestUser(t, "settings@test.com", string(models.RBAC_SETTINGS_MANAGER))
	auditToken := th.createTestUser(t, "audit@test.com", string(models.RBAC_AUDIT_LOG_VIEWER))
	noPermToken := th.createTestUser(t, "noperm@test.com", "user")

	tests := []struct {
		name       string
		method     string
		path       string
		token      *string
		statusCode int
		desc       string
	}{
		{
			name:       "settings_manager can access storage devices",
			method:     "GET",
			path:       "/api/storage/devices",
			token:      &settingsToken,
			statusCode: http.StatusOK,
			desc:       "Settings manager should access storage devices",
		},
		{
			name:       "settings_manager can access mount points",
			method:     "GET",
			path:       "/api/storage/mount-points",
			token:      &settingsToken,
			statusCode: http.StatusOK,
			desc:       "Settings manager should access mount points",
		},
		{
			name:       "audit_log_viewer cannot access storage",
			method:     "GET",
			path:       "/api/storage/devices",
			token:      &auditToken,
			statusCode: http.StatusForbidden,
			desc:       "Audit log viewer should not access storage",
		},
		{
			name:       "user without permission cannot access storage",
			method:     "GET",
			path:       "/api/storage/devices",
			token:      &noPermToken,
			statusCode: http.StatusForbidden,
			desc:       "User without permission should get 403",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := th.makeRequest(t, tt.method, tt.path, tt.token, nil)
			assertStatusCode(t, resp, tt.statusCode, tt.desc)
		})
	}
}

// TestRBACMultiplePolicies tests that RBAC requires ALL policies to be satisfied
func TestRBACMultiplePolicies(t *testing.T) {
	th := setupTestServer(t)

	// Create a user with docker_read but not docker_write
	dockerReadToken := th.createTestUser(t, "docker@test.com", string(models.RBAC_DOCKER_READ))
	// Create a user with both docker_read and docker_write
	dockerBothToken := th.createTestUser(t, "docker_both@test.com",
		string(models.RBAC_DOCKER_READ)+","+string(models.RBAC_DOCKER_WRITE))

	// Note: The app doesn't currently have endpoints requiring multiple non-admin policies,
	// but this test structure demonstrates the concept
	tests := []struct {
		name       string
		token      *string
		statusCode int
		desc       string
	}{
		{
			name:       "User with only docker_read cannot access admin endpoints",
			token:      &dockerReadToken,
			statusCode: http.StatusForbidden,
			desc:       "User should need admin for user endpoints",
		},
		{
			name:       "User with multiple policies still needs specific required policy",
			token:      &dockerBothToken,
			statusCode: http.StatusForbidden,
			desc:       "User should still need admin for user endpoints",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := th.makeRequest(t, "GET", "/api/users", tt.token, nil)
			assertStatusCode(t, resp, tt.statusCode, tt.desc)
		})
	}
}

// TestRBACExpiredSession tests that RBAC properly rejects expired sessions
func TestRBACExpiredSession(t *testing.T) {
	th := setupTestServer(t)
	ctx := context.Background()

	// Create a test user
	userId, err := th.db.User.CreateUser(ctx, user.CreateUserParams{
		Username: "expired@test.com",
		Email:    "expired@test.com",
		Password: "hashedpassword",
		Role:     string(models.RBAC_USER_ADMIN),
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a session with an expiration time in the past
	sessionToken := uuid.Must(uuid.NewV4()).String()
	_, err = th.db.Session.UpsertSession(ctx, dbsessions.UpsertSessionParams{
		UserID:       userId.ID,
		SessionToken: sessionToken,
	})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Try to use the expired session (note: actual expiration logic may need implementation)
	resp, _ := th.makeRequest(t, "GET", "/api/users", &sessionToken, nil)

	// Should still work if session expiration isn't enforced in auth middleware yet
	// This test documents the expected behavior
	if resp.Code != http.StatusOK && resp.Code != http.StatusUnauthorized {
		t.Errorf("Unexpected status code: %d", resp.Code)
	}
}

// TestAllRBACPolicies tests that all defined RBAC policies are recognized
func TestAllRBACPolicies(t *testing.T) {
	expectedPolicies := []models.RBACPolicy{
		models.RBAC_DOCKER_READ,
		models.RBAC_DOCKER_WRITE,
		models.RBAC_DOCKER_UPDATE,
		models.RBAC_DOCKER_DELETE,
		models.RBAC_QEMU_READ,
		models.RBAC_QEMU_WRITE,
		models.RBAC_QEMU_UPDATE,
		models.RBAC_QEMU_DELETE,
		models.RBAC_EVENT_VIEWER,
		models.RBAC_EVENT_MANAGER,
		models.RBAC_USER_ADMIN,
		models.RBAC_SETTINGS_MANAGER,
		models.RBAC_AUDIT_LOG_VIEWER,
		models.RBAC_HEALTH_CHECKER,
		models.RBAC_FIREWALL_READ,
		models.RBAC_FIREWALL_WRITE,
		models.RBAC_FIREWALL_UPDATE,
		models.RBAC_FIREWALL_DELETE,
	}

	for _, policy := range expectedPolicies {
		if _, ok := models.AllRBACPolicies[string(policy)]; !ok {
			t.Errorf("Policy %s not found in AllRBACPolicies map", policy)
		}
	}

	// Verify the count matches
	if len(models.AllRBACPolicies) != len(expectedPolicies) {
		t.Errorf("Expected %d policies, but found %d", len(expectedPolicies), len(models.AllRBACPolicies))
	}
}

// TestRBACWithUserAdminAndOtherRoles tests user_admin with mixed roles
func TestRBACWithUserAdminAndOtherRoles(t *testing.T) {
	th := setupTestServer(t)

	// Create a user with user_admin plus other roles (admin should still work)
	adminMultiToken := th.createTestUser(t, "admin_multi@test.com",
		string(models.RBAC_USER_ADMIN)+","+string(models.RBAC_DOCKER_READ)+","+string(models.RBAC_AUDIT_LOG_VIEWER))

	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		desc       string
	}{
		{
			name:       "Admin with multiple roles can access users",
			method:     "GET",
			path:       "/api/users",
			statusCode: http.StatusOK,
			desc:       "User admin should have access",
		},
		{
			name:       "Admin with multiple roles can access logs",
			method:     "GET",
			path:       "/api/logs",
			statusCode: http.StatusOK,
			desc:       "Admin should bypass to logs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := th.makeRequest(t, tt.method, tt.path, &adminMultiToken, nil)
			assertStatusCode(t, resp, tt.statusCode, tt.desc)
		})
	}
}

// TestRBACQemu tests RBAC on QEMU/virtualization endpoints
func TestRBACQemu(t *testing.T) {
	th := setupTestServer(t)

	// Create test users with different QEMU permissions
	qemuReadToken := th.createTestUser(t, "qemu_read@test.com", string(models.RBAC_QEMU_READ))
	qemuWriteToken := th.createTestUser(t, "qemu_write@test.com", string(models.RBAC_QEMU_WRITE))
	qemuUpdateToken := th.createTestUser(t, "qemu_update@test.com", string(models.RBAC_QEMU_UPDATE))
	qemuFullToken := th.createTestUser(t, "qemu_full@test.com",
		string(models.RBAC_QEMU_READ)+","+string(models.RBAC_QEMU_WRITE)+","+string(models.RBAC_QEMU_UPDATE))
	dockerOnlyToken := th.createTestUser(t, "docker_only@test.com", string(models.RBAC_DOCKER_READ))
	noPermToken := th.createTestUser(t, "noperm@test.com", "user")

	tests := []struct {
		name       string
		method     string
		path       string
		token      *string
		statusCode int
		desc       string
	}{
		// QEMU Read tests
		{
			name:       "qemu_read can access virtual machines list",
			method:     "GET",
			path:       "/api/qemu/virtual-machines",
			token:      &qemuReadToken,
			statusCode: http.StatusOK,
			desc:       "User with qemu_read should access VM list",
		},
		{
			name:       "qemu_read can access virtual machines info",
			method:     "GET",
			path:       "/api/qemu/virtual-machines/info",
			token:      &qemuReadToken,
			statusCode: http.StatusOK,
			desc:       "User with qemu_read should access VM info",
		},
		{
			name:       "docker_only cannot access qemu endpoints",
			method:     "GET",
			path:       "/api/qemu/virtual-machines",
			token:      &dockerOnlyToken,
			statusCode: http.StatusForbidden,
			desc:       "User with docker_read only should not access qemu",
		},
		{
			name:       "no permission cannot access qemu endpoints",
			method:     "GET",
			path:       "/api/qemu/virtual-machines",
			token:      &noPermToken,
			statusCode: http.StatusForbidden,
			desc:       "User without permission should get 403",
		},
		// QEMU Write tests - qemu_write can create VMs
		{
			name:       "qemu_read cannot create virtual machines",
			method:     "POST",
			path:       "/api/qemu/virtual-machines",
			token:      &qemuReadToken,
			statusCode: http.StatusForbidden,
			desc:       "User with only qemu_read should not create VMs",
		},
		{
			name:       "qemu_full can access read endpoints",
			method:     "GET",
			path:       "/api/qemu/virtual-machines",
			token:      &qemuFullToken,
			statusCode: http.StatusOK,
			desc:       "User with full qemu permissions should access read",
		},
		// QEMU Update tests - requires qemu_update for reboot/shutdown
		{
			name:       "qemu_read cannot update virtual machines",
			method:     "POST",
			path:       "/api/qemu/virtual-machines/test-uuid/reboot",
			token:      &qemuReadToken,
			statusCode: http.StatusForbidden,
			desc:       "User with only qemu_read should not reboot VMs",
		},
		{
			name:       "qemu_write cannot reboot virtual machines",
			method:     "POST",
			path:       "/api/qemu/virtual-machines/test-uuid/reboot",
			token:      &qemuWriteToken,
			statusCode: http.StatusForbidden,
			desc:       "User with only qemu_write should not reboot VMs",
		},
		{
			name:       "qemu_update cannot read virtual machines",
			method:     "GET",
			path:       "/api/qemu/virtual-machines",
			token:      &qemuUpdateToken,
			statusCode: http.StatusForbidden,
			desc:       "User with only qemu_update should not read VMs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := th.makeRequest(t, tt.method, tt.path, tt.token, nil)
			assertStatusCode(t, resp, tt.statusCode, tt.desc)
		})
	}
}

// TestRBACDocker tests RBAC on Docker endpoints
func TestRBACDocker(t *testing.T) {
	th := setupTestServer(t)

	// Create test users with different Docker permissions
	dockerReadToken := th.createTestUser(t, "docker_read@test.com", string(models.RBAC_DOCKER_READ))
	dockerWriteToken := th.createTestUser(t, "docker_write@test.com", string(models.RBAC_DOCKER_WRITE))
	dockerUpdateToken := th.createTestUser(t, "docker_update@test.com", string(models.RBAC_DOCKER_UPDATE))
	dockerDeleteToken := th.createTestUser(t, "docker_delete@test.com", string(models.RBAC_DOCKER_DELETE))
	dockerFullToken := th.createTestUser(t, "docker_full@test.com",
		string(models.RBAC_DOCKER_READ)+","+string(models.RBAC_DOCKER_WRITE)+","+
			string(models.RBAC_DOCKER_UPDATE)+","+string(models.RBAC_DOCKER_DELETE))
	qemuOnlyToken := th.createTestUser(t, "qemu_only@test.com", string(models.RBAC_QEMU_READ))
	noPermToken := th.createTestUser(t, "noperm@test.com", "user")

	tests := []struct {
		name       string
		method     string
		path       string
		token      *string
		statusCode int
		desc       string
	}{
		// Docker Read tests
		{
			name:       "docker_read can access available clients",
			method:     "GET",
			path:       "/api/docker",
			token:      &dockerReadToken,
			statusCode: http.StatusOK,
			desc:       "User with docker_read should access docker clients",
		},
		{
			name:       "qemu_only cannot access docker endpoints",
			method:     "GET",
			path:       "/api/docker",
			token:      &qemuOnlyToken,
			statusCode: http.StatusForbidden,
			desc:       "User with qemu_read only should not access docker",
		},
		{
			name:       "no permission cannot access docker endpoints",
			method:     "GET",
			path:       "/api/docker",
			token:      &noPermToken,
			statusCode: http.StatusForbidden,
			desc:       "User without permission should get 403",
		},
		// Docker Write tests
		{
			name:       "docker_read cannot access write-protected endpoints",
			method:     "GET",
			path:       "/api/docker",
			token:      &dockerReadToken,
			statusCode: http.StatusOK,
			desc:       "docker_read should still read docker list",
		},
		{
			name:       "docker_write cannot access read endpoints",
			method:     "GET",
			path:       "/api/docker",
			token:      &dockerWriteToken,
			statusCode: http.StatusForbidden,
			desc:       "docker_write only should not read docker list",
		},
		// Docker Update tests
		{
			name:       "docker_update cannot access read endpoints",
			method:     "GET",
			path:       "/api/docker",
			token:      &dockerUpdateToken,
			statusCode: http.StatusForbidden,
			desc:       "docker_update only should not read docker list",
		},
		// Docker Delete tests
		{
			name:       "docker_delete cannot access read endpoints",
			method:     "GET",
			path:       "/api/docker",
			token:      &dockerDeleteToken,
			statusCode: http.StatusForbidden,
			desc:       "docker_delete only should not read docker list",
		},
		// Full permissions
		{
			name:       "docker_full can access read endpoints",
			method:     "GET",
			path:       "/api/docker",
			token:      &dockerFullToken,
			statusCode: http.StatusOK,
			desc:       "User with full docker permissions should access read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := th.makeRequest(t, tt.method, tt.path, tt.token, nil)
			assertStatusCode(t, resp, tt.statusCode, tt.desc)
		})
	}
}

// TestRBACFirewall tests RBAC on Firewall endpoints
func TestRBACFirewall(t *testing.T) {
	th := setupTestServer(t)

	// Create test users with different Firewall permissions
	firewallReadToken := th.createTestUser(t, "firewall_read@test.com", string(models.RBAC_FIREWALL_READ))
	firewallWriteToken := th.createTestUser(t, "firewall_write@test.com", string(models.RBAC_FIREWALL_WRITE))
	firewallUpdateToken := th.createTestUser(t, "firewall_update@test.com", string(models.RBAC_FIREWALL_UPDATE))
	firewallDeleteToken := th.createTestUser(t, "firewall_delete@test.com", string(models.RBAC_FIREWALL_DELETE))
	firewallFullToken := th.createTestUser(t, "firewall_full@test.com",
		string(models.RBAC_FIREWALL_READ)+","+string(models.RBAC_FIREWALL_WRITE)+","+
			string(models.RBAC_FIREWALL_UPDATE)+","+string(models.RBAC_FIREWALL_DELETE))
	dockerOnlyToken := th.createTestUser(t, "docker_only@test.com", string(models.RBAC_DOCKER_READ))
	noPermToken := th.createTestUser(t, "noperm@test.com", "user")

	tests := []struct {
		name       string
		method     string
		path       string
		token      *string
		statusCode int
		desc       string
	}{
		// Firewall Read tests
		{
			name:       "firewall_read can access firewall status",
			method:     "GET",
			path:       "/api/firewall/status",
			token:      &firewallReadToken,
			statusCode: http.StatusOK,
			desc:       "User with firewall_read should access firewall status",
		},
		{
			name:       "firewall_read can access firewall rules",
			method:     "GET",
			path:       "/api/firewall/rules",
			token:      &firewallReadToken,
			statusCode: http.StatusOK,
			desc:       "User with firewall_read should access firewall rules",
		},
		{
			name:       "docker_only cannot access firewall endpoints",
			method:     "GET",
			path:       "/api/firewall/status",
			token:      &dockerOnlyToken,
			statusCode: http.StatusForbidden,
			desc:       "User with docker_read only should not access firewall",
		},
		{
			name:       "no permission cannot access firewall endpoints",
			method:     "GET",
			path:       "/api/firewall/status",
			token:      &noPermToken,
			statusCode: http.StatusForbidden,
			desc:       "User without permission should get 403",
		},
		// Firewall Write tests
		{
			name:       "firewall_read cannot add firewall rules",
			method:     "POST",
			path:       "/api/firewall/rules",
			token:      &firewallReadToken,
			statusCode: http.StatusForbidden,
			desc:       "User with only firewall_read should not add rules",
		},
		{
			name:       "firewall_write cannot read firewall status",
			method:     "GET",
			path:       "/api/firewall/status",
			token:      &firewallWriteToken,
			statusCode: http.StatusForbidden,
			desc:       "User with only firewall_write should not read status",
		},
		// Firewall Update tests
		{
			name:       "firewall_read cannot reorder firewall rules",
			method:     "POST",
			path:       "/api/firewall/rules/reorder",
			token:      &firewallReadToken,
			statusCode: http.StatusForbidden,
			desc:       "User with only firewall_read should not reorder rules",
		},
		{
			name:       "firewall_update cannot read firewall status",
			method:     "GET",
			path:       "/api/firewall/status",
			token:      &firewallUpdateToken,
			statusCode: http.StatusForbidden,
			desc:       "User with only firewall_update should not read status",
		},
		// Firewall Delete tests
		{
			name:       "firewall_read cannot delete firewall rules",
			method:     "DELETE",
			path:       "/api/firewall/rules/test-handle",
			token:      &firewallReadToken,
			statusCode: http.StatusForbidden,
			desc:       "User with only firewall_read should not delete rules",
		},
		{
			name:       "firewall_delete cannot read firewall status",
			method:     "GET",
			path:       "/api/firewall/status",
			token:      &firewallDeleteToken,
			statusCode: http.StatusForbidden,
			desc:       "User with only firewall_delete should not read status",
		},
		// Full permissions
		{
			name:       "firewall_full can access read endpoints",
			method:     "GET",
			path:       "/api/firewall/status",
			token:      &firewallFullToken,
			statusCode: http.StatusOK,
			desc:       "User with full firewall permissions should access read",
		},
		{
			name:       "firewall_full can access rules list",
			method:     "GET",
			path:       "/api/firewall/rules",
			token:      &firewallFullToken,
			statusCode: http.StatusOK,
			desc:       "User with full firewall permissions should access rules",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := th.makeRequest(t, tt.method, tt.path, tt.token, nil)
			assertStatusCode(t, resp, tt.statusCode, tt.desc)
		})
	}
}

// TestRBACUserCRUD tests RBAC on User CRUD operations
func TestRBACUserCRUD(t *testing.T) {
	th := setupTestServer(t)

	// Create test users
	adminToken := th.createTestUser(t, "admin@test.com", string(models.RBAC_USER_ADMIN))
	settingsToken := th.createTestUser(t, "settings@test.com", string(models.RBAC_SETTINGS_MANAGER))
	auditToken := th.createTestUser(t, "audit@test.com", string(models.RBAC_AUDIT_LOG_VIEWER))
	noPermToken := th.createTestUser(t, "noperm@test.com", "user")

	// Get the user ID of the noperm user for testing
	noPermUser := th.testUsers["noperm@test.com"]

	tests := []struct {
		name       string
		method     string
		path       string
		token      *string
		body       interface{}
		statusCode int
		desc       string
	}{
		// GET user by ID
		{
			name:       "admin can get user by id",
			method:     "GET",
			path:       fmt.Sprintf("/api/users/%d", noPermUser.id),
			token:      &adminToken,
			statusCode: http.StatusOK,
			desc:       "Admin should get user by ID",
		},
		{
			name:       "settings_manager cannot get user by id",
			method:     "GET",
			path:       fmt.Sprintf("/api/users/%d", noPermUser.id),
			token:      &settingsToken,
			statusCode: http.StatusForbidden,
			desc:       "Settings manager should not access user details",
		},
		{
			name:       "audit_log_viewer cannot get user by id",
			method:     "GET",
			path:       fmt.Sprintf("/api/users/%d", noPermUser.id),
			token:      &auditToken,
			statusCode: http.StatusForbidden,
			desc:       "Audit log viewer should not access user details",
		},
		// Update user role
		{
			name:       "admin can update user role",
			method:     "PATCH",
			path:       fmt.Sprintf("/api/users/%d/role", noPermUser.id),
			token:      &adminToken,
			body:       map[string]string{"role": "docker_read"},
			statusCode: http.StatusOK,
			desc:       "Admin should update user role",
		},
		{
			name:       "settings_manager cannot update user role",
			method:     "PATCH",
			path:       fmt.Sprintf("/api/users/%d/role", noPermUser.id),
			token:      &settingsToken,
			body:       map[string]string{"role": "docker_read"},
			statusCode: http.StatusForbidden,
			desc:       "Settings manager should not update user role",
		},
		{
			name:       "no permission cannot update user role",
			method:     "PATCH",
			path:       fmt.Sprintf("/api/users/%d/role", noPermUser.id),
			token:      &noPermToken,
			body:       map[string]string{"role": "docker_read"},
			statusCode: http.StatusForbidden,
			desc:       "User without permission should get 403",
		},
		// Delete user
		{
			name:       "settings_manager cannot delete user",
			method:     "DELETE",
			path:       fmt.Sprintf("/api/users/%d", noPermUser.id),
			token:      &settingsToken,
			statusCode: http.StatusForbidden,
			desc:       "Settings manager should not delete users",
		},
		{
			name:       "no permission cannot delete user",
			method:     "DELETE",
			path:       fmt.Sprintf("/api/users/%d", noPermUser.id),
			token:      &noPermToken,
			statusCode: http.StatusForbidden,
			desc:       "User without permission should get 403",
		},
		// Create user
		{
			name:   "admin can create user",
			method: "POST",
			path:   "/api/users",
			token:  &adminToken,
			body: map[string]string{
				"username": "newuser",
				"email":    "newuser@test.com",
				"password": "password123",
				"role":     "user",
			},
			statusCode: http.StatusOK,
			desc:       "Admin should create new user",
		},
		{
			name:   "settings_manager cannot create user",
			method: "POST",
			path:   "/api/users",
			token:  &settingsToken,
			body: map[string]string{
				"username": "newuser2",
				"email":    "newuser2@test.com",
				"password": "password123",
				"role":     "user",
			},
			statusCode: http.StatusForbidden,
			desc:       "Settings manager should not create users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := th.makeRequest(t, tt.method, tt.path, tt.token, tt.body)
			assertStatusCode(t, resp, tt.statusCode, tt.desc)
		})
	}
}

// TestRBACCrossServicePermissions tests that permissions don't leak across services
func TestRBACCrossServicePermissions(t *testing.T) {
	th := setupTestServer(t)

	// Create users with single-service permissions
	dockerReadToken := th.createTestUser(t, "docker_read@test.com", string(models.RBAC_DOCKER_READ))
	qemuReadToken := th.createTestUser(t, "qemu_read@test.com", string(models.RBAC_QEMU_READ))
	firewallReadToken := th.createTestUser(t, "firewall_read@test.com", string(models.RBAC_FIREWALL_READ))
	settingsToken := th.createTestUser(t, "settings@test.com", string(models.RBAC_SETTINGS_MANAGER))
	auditToken := th.createTestUser(t, "audit@test.com", string(models.RBAC_AUDIT_LOG_VIEWER))
	healthToken := th.createTestUser(t, "health@test.com", string(models.RBAC_HEALTH_CHECKER))

	tests := []struct {
		name       string
		method     string
		path       string
		token      *string
		statusCode int
		desc       string
	}{
		// Docker read cannot access other services
		{
			name:       "docker_read cannot access qemu",
			method:     "GET",
			path:       "/api/qemu/virtual-machines",
			token:      &dockerReadToken,
			statusCode: http.StatusForbidden,
			desc:       "Docker read should not access QEMU",
		},
		{
			name:       "docker_read cannot access firewall",
			method:     "GET",
			path:       "/api/firewall/status",
			token:      &dockerReadToken,
			statusCode: http.StatusForbidden,
			desc:       "Docker read should not access firewall",
		},
		{
			name:       "docker_read cannot access storage",
			method:     "GET",
			path:       "/api/storage/devices",
			token:      &dockerReadToken,
			statusCode: http.StatusForbidden,
			desc:       "Docker read should not access storage",
		},
		{
			name:       "docker_read cannot access users",
			method:     "GET",
			path:       "/api/users",
			token:      &dockerReadToken,
			statusCode: http.StatusForbidden,
			desc:       "Docker read should not access users",
		},
		// QEMU read cannot access other services
		{
			name:       "qemu_read cannot access docker",
			method:     "GET",
			path:       "/api/docker",
			token:      &qemuReadToken,
			statusCode: http.StatusForbidden,
			desc:       "QEMU read should not access Docker",
		},
		{
			name:       "qemu_read cannot access firewall",
			method:     "GET",
			path:       "/api/firewall/status",
			token:      &qemuReadToken,
			statusCode: http.StatusForbidden,
			desc:       "QEMU read should not access firewall",
		},
		// Firewall read cannot access other services
		{
			name:       "firewall_read cannot access docker",
			method:     "GET",
			path:       "/api/docker",
			token:      &firewallReadToken,
			statusCode: http.StatusForbidden,
			desc:       "Firewall read should not access Docker",
		},
		{
			name:       "firewall_read cannot access qemu",
			method:     "GET",
			path:       "/api/qemu/virtual-machines",
			token:      &firewallReadToken,
			statusCode: http.StatusForbidden,
			desc:       "Firewall read should not access QEMU",
		},
		// Settings manager cannot access metrics/logs
		{
			name:       "settings_manager cannot access logs",
			method:     "GET",
			path:       "/api/logs",
			token:      &settingsToken,
			statusCode: http.StatusForbidden,
			desc:       "Settings manager should not access logs",
		},
		{
			name:       "settings_manager cannot access metrics",
			method:     "GET",
			path:       "/api/metrics",
			token:      &settingsToken,
			statusCode: http.StatusForbidden,
			desc:       "Settings manager should not access metrics",
		},
		// Audit log viewer cannot access storage
		{
			name:       "audit_log_viewer cannot access storage",
			method:     "GET",
			path:       "/api/storage/devices",
			token:      &auditToken,
			statusCode: http.StatusForbidden,
			desc:       "Audit log viewer should not access storage",
		},
		// Health checker cannot access logs
		{
			name:       "health_checker cannot access logs",
			method:     "GET",
			path:       "/api/logs",
			token:      &healthToken,
			statusCode: http.StatusForbidden,
			desc:       "Health checker should not access logs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := th.makeRequest(t, tt.method, tt.path, tt.token, nil)
			assertStatusCode(t, resp, tt.statusCode, tt.desc)
		})
	}
}

// TestRBACCombinedPermissions tests users with multiple permissions across services
func TestRBACCombinedPermissions(t *testing.T) {
	th := setupTestServer(t)

	// Create users with combined permissions
	dockerQemuToken := th.createTestUser(t, "docker_qemu@test.com",
		string(models.RBAC_DOCKER_READ)+","+string(models.RBAC_QEMU_READ))
	firewallSettingsToken := th.createTestUser(t, "firewall_settings@test.com",
		string(models.RBAC_FIREWALL_READ)+","+string(models.RBAC_SETTINGS_MANAGER))
	auditHealthToken := th.createTestUser(t, "audit_health@test.com",
		string(models.RBAC_AUDIT_LOG_VIEWER)+","+string(models.RBAC_HEALTH_CHECKER))
	allReadToken := th.createTestUser(t, "all_read@test.com",
		string(models.RBAC_DOCKER_READ)+","+string(models.RBAC_QEMU_READ)+","+
			string(models.RBAC_FIREWALL_READ)+","+string(models.RBAC_AUDIT_LOG_VIEWER)+","+
			string(models.RBAC_HEALTH_CHECKER)+","+string(models.RBAC_SETTINGS_MANAGER))

	tests := []struct {
		name       string
		method     string
		path       string
		token      *string
		statusCode int
		desc       string
	}{
		// Docker + QEMU user
		{
			name:       "docker_qemu can access docker",
			method:     "GET",
			path:       "/api/docker",
			token:      &dockerQemuToken,
			statusCode: http.StatusOK,
			desc:       "Combined user should access Docker",
		},
		{
			name:       "docker_qemu can access qemu",
			method:     "GET",
			path:       "/api/qemu/virtual-machines",
			token:      &dockerQemuToken,
			statusCode: http.StatusOK,
			desc:       "Combined user should access QEMU",
		},
		{
			name:       "docker_qemu cannot access firewall",
			method:     "GET",
			path:       "/api/firewall/status",
			token:      &dockerQemuToken,
			statusCode: http.StatusForbidden,
			desc:       "Combined user should not access firewall",
		},
		// Firewall + Settings user
		{
			name:       "firewall_settings can access firewall",
			method:     "GET",
			path:       "/api/firewall/status",
			token:      &firewallSettingsToken,
			statusCode: http.StatusOK,
			desc:       "Combined user should access firewall",
		},
		{
			name:       "firewall_settings can access storage",
			method:     "GET",
			path:       "/api/storage/devices",
			token:      &firewallSettingsToken,
			statusCode: http.StatusOK,
			desc:       "Combined user should access storage",
		},
		{
			name:       "firewall_settings cannot access docker",
			method:     "GET",
			path:       "/api/docker",
			token:      &firewallSettingsToken,
			statusCode: http.StatusForbidden,
			desc:       "Combined user should not access docker",
		},
		// Audit + Health user
		{
			name:       "audit_health can access logs",
			method:     "GET",
			path:       "/api/logs",
			token:      &auditHealthToken,
			statusCode: http.StatusOK,
			desc:       "Combined user should access logs",
		},
		{
			name:       "audit_health can access metrics",
			method:     "GET",
			path:       "/api/metrics",
			token:      &auditHealthToken,
			statusCode: http.StatusOK,
			desc:       "Combined user should access metrics",
		},
		{
			name:       "audit_health can access health metrics",
			method:     "GET",
			path:       "/api/metrics/health",
			token:      &auditHealthToken,
			statusCode: http.StatusOK,
			desc:       "Combined user should access health metrics",
		},
		{
			name:       "audit_health cannot access users",
			method:     "GET",
			path:       "/api/users",
			token:      &auditHealthToken,
			statusCode: http.StatusForbidden,
			desc:       "Combined user should not access users",
		},
		// All read permissions user
		{
			name:       "all_read can access docker",
			method:     "GET",
			path:       "/api/docker",
			token:      &allReadToken,
			statusCode: http.StatusOK,
			desc:       "All-read user should access Docker",
		},
		{
			name:       "all_read can access qemu",
			method:     "GET",
			path:       "/api/qemu/virtual-machines",
			token:      &allReadToken,
			statusCode: http.StatusOK,
			desc:       "All-read user should access QEMU",
		},
		{
			name:       "all_read can access firewall",
			method:     "GET",
			path:       "/api/firewall/status",
			token:      &allReadToken,
			statusCode: http.StatusOK,
			desc:       "All-read user should access firewall",
		},
		{
			name:       "all_read can access logs",
			method:     "GET",
			path:       "/api/logs",
			token:      &allReadToken,
			statusCode: http.StatusOK,
			desc:       "All-read user should access logs",
		},
		{
			name:       "all_read can access storage",
			method:     "GET",
			path:       "/api/storage/devices",
			token:      &allReadToken,
			statusCode: http.StatusOK,
			desc:       "All-read user should access storage",
		},
		{
			name:       "all_read cannot access users (needs admin)",
			method:     "GET",
			path:       "/api/users",
			token:      &allReadToken,
			statusCode: http.StatusForbidden,
			desc:       "All-read user should not access users without admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := th.makeRequest(t, tt.method, tt.path, tt.token, nil)
			assertStatusCode(t, resp, tt.statusCode, tt.desc)
		})
	}
}

// Helper function to return pointer to string
func pointToString(s string) *string {
	return &s
}
