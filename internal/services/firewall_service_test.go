package services

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"visory/internal/models"
	"visory/internal/utils"

	"github.com/google/nftables"
)

// testTableName returns a unique table name for testing
func testTableName() string {
	return fmt.Sprintf("visory_test_%d", time.Now().UnixNano())
}

// skipIfNotRoot skips the test if not running as root
func skipIfNotRoot(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("Skipping test: requires root privileges (run with sudo)")
	}
}

// cleanupTestTable removes a test table if it exists
func cleanupTestTable(conn *nftables.Conn, tableName string) {
	tables, err := conn.ListTables()
	if err != nil {
		return
	}
	for _, t := range tables {
		if t.Name == tableName {
			conn.DelTable(t)
			conn.Flush()
			return
		}
	}
}

func TestFirewallService_Initialize(t *testing.T) {
	skipIfNotRoot(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dispatcher := &utils.Dispatcher{}

	service := NewFirewallService(dispatcher, logger)

	if service == nil {
		t.Fatal("expected service to be non-nil")
	}

	if service.conn == nil {
		t.Fatal("expected nftables connection to be initialized")
	}

	if service.table == nil {
		t.Fatal("expected table to be initialized")
	}

	if len(service.chains) == 0 {
		t.Fatal("expected chains to be initialized")
	}

	// Verify chains exist
	expectedChains := []string{"input", "forward", "output"}
	for _, chainName := range expectedChains {
		if _, ok := service.chains[chainName]; !ok {
			t.Errorf("expected chain %s to exist", chainName)
		}
	}
}

func TestFirewallService_AddAndDeleteRule(t *testing.T) {
	skipIfNotRoot(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dispatcher := &utils.Dispatcher{}

	service := NewFirewallService(dispatcher, logger)
	if service == nil || service.conn == nil {
		t.Fatal("failed to initialize firewall service")
	}

	// Count initial rules
	initialCount := 0
	for _, chain := range service.chains {
		rules, err := service.conn.GetRules(service.table, chain)
		if err == nil {
			initialCount += len(rules)
		}
	}

	// Create a test rule
	req := models.CreateRuleRequest{
		Chain:    "input",
		Protocol: "tcp",
		Port:     12345,
		Action:   "accept",
		Comment:  "test rule",
	}

	exprs := service.buildRuleExprs(req)
	chain := service.chains["input"]

	rule := &nftables.Rule{
		Table: service.table,
		Chain: chain,
		Exprs: exprs,
	}

	service.conn.AddRule(rule)
	if err := service.conn.Flush(); err != nil {
		t.Fatalf("failed to add rule: %v", err)
	}

	// Verify rule was added
	rules, err := service.conn.GetRules(service.table, chain)
	if err != nil {
		t.Fatalf("failed to get rules: %v", err)
	}

	if len(rules) <= initialCount {
		t.Error("expected rule count to increase after adding rule")
	}

	// Get the last rule (the one we just added)
	if len(rules) == 0 {
		t.Fatal("no rules found after adding")
	}
	lastRule := rules[len(rules)-1]

	// Delete the rule
	service.conn.DelRule(&nftables.Rule{
		Table:  service.table,
		Chain:  chain,
		Handle: lastRule.Handle,
	})

	if err := service.conn.Flush(); err != nil {
		t.Fatalf("failed to delete rule: %v", err)
	}

	// Verify rule was deleted
	rulesAfter, err := service.conn.GetRules(service.table, chain)
	if err != nil {
		t.Fatalf("failed to get rules after delete: %v", err)
	}

	if len(rulesAfter) >= len(rules) {
		t.Error("expected rule count to decrease after deleting rule")
	}
}

func TestFirewallService_BuildRuleExprs_TCP(t *testing.T) {
	skipIfNotRoot(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dispatcher := &utils.Dispatcher{}

	service := NewFirewallService(dispatcher, logger)
	if service == nil {
		t.Fatal("failed to initialize firewall service")
	}

	req := models.CreateRuleRequest{
		Chain:    "input",
		Protocol: "tcp",
		Port:     443,
		Action:   "accept",
	}

	exprs := service.buildRuleExprs(req)

	if len(exprs) == 0 {
		t.Fatal("expected expressions to be generated")
	}

	// Should have: protocol match (2), port match (2), counter (1), verdict (1) = 6 exprs
	expectedCount := 6
	if len(exprs) != expectedCount {
		t.Errorf("expected %d expressions, got %d", expectedCount, len(exprs))
	}
}

func TestFirewallService_BuildRuleExprs_UDP(t *testing.T) {
	skipIfNotRoot(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dispatcher := &utils.Dispatcher{}

	service := NewFirewallService(dispatcher, logger)
	if service == nil {
		t.Fatal("failed to initialize firewall service")
	}

	req := models.CreateRuleRequest{
		Chain:    "input",
		Protocol: "udp",
		Port:     53,
		Action:   "drop",
	}

	exprs := service.buildRuleExprs(req)

	if len(exprs) == 0 {
		t.Fatal("expected expressions to be generated")
	}
}

func TestFirewallService_BuildRuleExprs_WithSourceIP(t *testing.T) {
	skipIfNotRoot(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dispatcher := &utils.Dispatcher{}

	service := NewFirewallService(dispatcher, logger)
	if service == nil {
		t.Fatal("failed to initialize firewall service")
	}

	req := models.CreateRuleRequest{
		Chain:    "input",
		Protocol: "tcp",
		Port:     22,
		SourceIP: "192.168.1.100",
		Action:   "accept",
	}

	exprs := service.buildRuleExprs(req)

	if len(exprs) == 0 {
		t.Fatal("expected expressions to be generated")
	}

	// Should have: protocol (2), source IP (2), port (2), counter (1), verdict (1) = 8 exprs
	expectedCount := 8
	if len(exprs) != expectedCount {
		t.Errorf("expected %d expressions, got %d", expectedCount, len(exprs))
	}
}

func TestFirewallService_BuildRuleExprs_SimpleAction(t *testing.T) {
	skipIfNotRoot(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dispatcher := &utils.Dispatcher{}

	service := NewFirewallService(dispatcher, logger)
	if service == nil {
		t.Fatal("failed to initialize firewall service")
	}

	// Test a simple drop all rule
	req := models.CreateRuleRequest{
		Chain:  "input",
		Action: "drop",
	}

	exprs := service.buildRuleExprs(req)

	// Should have: counter (1), verdict (1) = 2 exprs
	expectedCount := 2
	if len(exprs) != expectedCount {
		t.Errorf("expected %d expressions, got %d", expectedCount, len(exprs))
	}
}

func TestFirewallService_ParseRule(t *testing.T) {
	skipIfNotRoot(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dispatcher := &utils.Dispatcher{}

	service := NewFirewallService(dispatcher, logger)
	if service == nil || service.conn == nil {
		t.Fatal("failed to initialize firewall service")
	}

	// Add a test rule
	req := models.CreateRuleRequest{
		Chain:    "input",
		Protocol: "tcp",
		Port:     8080,
		Action:   "accept",
	}

	exprs := service.buildRuleExprs(req)
	chain := service.chains["input"]

	rule := &nftables.Rule{
		Table: service.table,
		Chain: chain,
		Exprs: exprs,
	}

	service.conn.AddRule(rule)
	if err := service.conn.Flush(); err != nil {
		t.Fatalf("failed to add rule: %v", err)
	}

	// Get and parse the rule
	rules, err := service.conn.GetRules(service.table, chain)
	if err != nil {
		t.Fatalf("failed to get rules: %v", err)
	}

	if len(rules) == 0 {
		t.Fatal("no rules found")
	}

	lastRule := rules[len(rules)-1]
	parsed := service.parseRule(lastRule, "input")

	if parsed == nil {
		t.Fatal("failed to parse rule")
	}

	if parsed.Chain != "input" {
		t.Errorf("expected chain 'input', got '%s'", parsed.Chain)
	}

	if parsed.Action != "accept" {
		t.Errorf("expected action 'accept', got '%s'", parsed.Action)
	}

	// Cleanup
	service.conn.DelRule(&nftables.Rule{
		Table:  service.table,
		Chain:  chain,
		Handle: lastRule.Handle,
	})
	service.conn.Flush()
}

func TestFirewallService_MultipleChains(t *testing.T) {
	skipIfNotRoot(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dispatcher := &utils.Dispatcher{}

	service := NewFirewallService(dispatcher, logger)
	if service == nil || service.conn == nil {
		t.Fatal("failed to initialize firewall service")
	}

	// Add rules to different chains
	chains := []string{"input", "forward", "output"}
	addedRules := make(map[string]uint64)

	for _, chainName := range chains {
		req := models.CreateRuleRequest{
			Chain:    chainName,
			Protocol: "tcp",
			Port:     9999,
			Action:   "drop",
		}

		exprs := service.buildRuleExprs(req)
		chain := service.chains[chainName]

		rule := &nftables.Rule{
			Table: service.table,
			Chain: chain,
			Exprs: exprs,
		}

		service.conn.AddRule(rule)
		if err := service.conn.Flush(); err != nil {
			t.Fatalf("failed to add rule to %s chain: %v", chainName, err)
		}

		// Get the handle
		rules, _ := service.conn.GetRules(service.table, chain)
		if len(rules) > 0 {
			addedRules[chainName] = rules[len(rules)-1].Handle
		}
	}

	// Verify all rules exist
	for chainName, handle := range addedRules {
		chain := service.chains[chainName]
		rules, err := service.conn.GetRules(service.table, chain)
		if err != nil {
			t.Fatalf("failed to get rules from %s: %v", chainName, err)
		}

		found := false
		for _, r := range rules {
			if r.Handle == handle {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("rule not found in %s chain", chainName)
		}

		// Cleanup
		service.conn.DelRule(&nftables.Rule{
			Table:  service.table,
			Chain:  chain,
			Handle: handle,
		})
	}

	service.conn.Flush()
}
