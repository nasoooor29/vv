package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"visory/internal/models"
	"visory/internal/utils"

	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"github.com/labstack/echo/v4"
	"golang.org/x/sys/unix"
)

const (
	tableName       = "filter"
	persistenceFile = "firewall.json"
	tableFamily     = nftables.TableFamilyINet
	// Safety rule comment prefix to identify system-managed rules
	safetyRulePrefix = "__visory_safety__"
)

type FirewallService struct {
	Dispatcher     *utils.Dispatcher
	Logger         *slog.Logger
	mu             sync.RWMutex
	conn           *nftables.Conn
	table          *nftables.Table
	chains         map[string]*nftables.Chain
	dataDir        string
	onRulesChanged func(rules []models.FirewallRule) // Callback for when rules change (for backup)
}

// NewFirewallService creates a new FirewallService with dependency injection
func NewFirewallService(dispatcher *utils.Dispatcher, logger *slog.Logger) *FirewallService {
	service := &FirewallService{
		Dispatcher: dispatcher.WithGroup("firewall"),
		Logger:     logger.WithGroup("firewall"),
		chains:     make(map[string]*nftables.Chain),
		dataDir:    models.ENV_VARS.Directory,
	}

	if err := service.initialize(); err != nil {
		logger.Error("failed to initialize firewall service", "error", err)
	}

	return service
}

// SetOnRulesChanged sets a callback function that is called whenever firewall rules change
// This is used by the backup service to create backups of firewall rules
func (s *FirewallService) SetOnRulesChanged(callback func(rules []models.FirewallRule)) {
	s.onRulesChanged = callback
}

// initialize sets up nftables connection and ensures table/chains exist
func (s *FirewallService) initialize() error {
	conn, err := nftables.New()
	if err != nil {
		return fmt.Errorf("failed to create nftables connection: %w", err)
	}
	s.conn = conn

	// Try to find existing filter table
	tables, err := conn.ListTables()
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}

	for _, t := range tables {
		if t.Name == tableName && t.Family == tableFamily {
			s.table = t
			break
		}
	}

	// Create table if it doesn't exist
	if s.table == nil {
		s.table = conn.AddTable(&nftables.Table{
			Family: tableFamily,
			Name:   tableName,
		})
		if err := conn.Flush(); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// Ensure chains exist
	if err := s.ensureChains(); err != nil {
		return err
	}

	// Add essential safety rules (loopback, established connections)
	if err := s.ensureSafetyRules(); err != nil {
		s.Logger.Warn("failed to add safety rules", "error", err)
	}

	// Load persisted rules
	if err := s.loadPersistedRules(); err != nil {
		s.Logger.Warn("failed to load persisted rules", "error", err)
	}

	return nil
}

// ensureChains creates input, forward, output chains if they don't exist
func (s *FirewallService) ensureChains() error {
	chains, err := s.conn.ListChains()
	if err != nil {
		return fmt.Errorf("failed to list chains: %w", err)
	}

	// Map existing chains
	existingChains := make(map[string]*nftables.Chain)
	for _, c := range chains {
		if c.Table.Name == tableName && c.Table.Family == tableFamily {
			existingChains[c.Name] = c
		}
	}

	chainConfigs := []struct {
		name    string
		hooknum *nftables.ChainHook
	}{
		{"input", nftables.ChainHookInput},
		{"forward", nftables.ChainHookForward},
		{"output", nftables.ChainHookOutput},
	}

	policyAccept := nftables.ChainPolicyAccept

	for _, cfg := range chainConfigs {
		if existing, ok := existingChains[cfg.name]; ok {
			s.chains[cfg.name] = existing
		} else {
			chain := s.conn.AddChain(&nftables.Chain{
				Name:     cfg.name,
				Table:    s.table,
				Type:     nftables.ChainTypeFilter,
				Hooknum:  cfg.hooknum,
				Priority: nftables.ChainPriorityFilter,
				Policy:   &policyAccept,
			})
			s.chains[cfg.name] = chain
		}
	}

	if err := s.conn.Flush(); err != nil {
		return fmt.Errorf("failed to create chains: %w", err)
	}

	return nil
}

// ensureSafetyRules adds essential rules to prevent lockout
// These rules allow loopback traffic and established/related connections
func (s *FirewallService) ensureSafetyRules() error {
	inputChain := s.chains["input"]
	if inputChain == nil {
		return fmt.Errorf("input chain not found")
	}

	// Check if safety rules already exist
	rules, err := s.conn.GetRules(s.table, inputChain)
	if err != nil {
		return fmt.Errorf("failed to get rules: %w", err)
	}

	// Check if we already have safety rules by looking at the first rules
	hasSafetyRules := false
	for _, rule := range rules {
		// Check if this is an accept rule on loopback (our safety rule marker)
		for _, e := range rule.Exprs {
			if meta, ok := e.(*expr.Meta); ok {
				if meta.Key == expr.MetaKeyIIFNAME {
					hasSafetyRules = true
					break
				}
			}
		}
		if hasSafetyRules {
			break
		}
	}

	if hasSafetyRules {
		s.Logger.Debug("safety rules already exist")
		return nil
	}

	s.Logger.Info("adding safety rules to prevent lockout")

	// Rule 1: Accept all loopback (lo) interface traffic
	// This is critical for local services to communicate
	s.conn.InsertRule(&nftables.Rule{
		Table: s.table,
		Chain: inputChain,
		Exprs: []expr.Any{
			// Match interface name "lo"
			&expr.Meta{Key: expr.MetaKeyIIFNAME, Register: 1},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte("lo\x00"), // null-terminated string
			},
			&expr.Counter{},
			&expr.Verdict{Kind: expr.VerdictAccept},
		},
	})

	// Rule 2: Accept established and related connections
	// This is CRITICAL - without this, existing TCP connections break immediately
	// when you add a drop rule
	s.conn.InsertRule(&nftables.Rule{
		Table: s.table,
		Chain: inputChain,
		Exprs: []expr.Any{
			// Load conntrack state
			&expr.Ct{Key: expr.CtKeySTATE, Register: 1},
			// Match established (bit 1) or related (bit 2)
			&expr.Bitwise{
				SourceRegister: 1,
				DestRegister:   1,
				Len:            4,
				Mask:           binaryutil.NativeEndian.PutUint32(expr.CtStateBitESTABLISHED | expr.CtStateBitRELATED),
				Xor:            binaryutil.NativeEndian.PutUint32(0),
			},
			&expr.Cmp{
				Op:       expr.CmpOpNeq,
				Register: 1,
				Data:     binaryutil.NativeEndian.PutUint32(0),
			},
			&expr.Counter{},
			&expr.Verdict{Kind: expr.VerdictAccept},
		},
	})

	if err := s.conn.Flush(); err != nil {
		return fmt.Errorf("failed to add safety rules: %w", err)
	}

	s.Logger.Info("safety rules added successfully")
	return nil
}

// isSafetyRule checks if a rule is a system-managed safety rule
// Safety rules use Meta IIFNAME (loopback) or Ct (conntrack state)
func (s *FirewallService) isSafetyRule(rule *nftables.Rule) bool {
	for _, e := range rule.Exprs {
		switch v := e.(type) {
		case *expr.Meta:
			// Loopback rule uses IIFNAME
			if v.Key == expr.MetaKeyIIFNAME {
				return true
			}
		case *expr.Ct:
			// Established/related connection rule uses conntrack
			if v.Key == expr.CtKeySTATE {
				return true
			}
		}
	}
	return false
}

// GetStatus returns the current firewall status
//
//	@Summary      get firewall status
//	@Description  returns the current state of the firewall
//	@Tags         firewall
//	@Produce      json
//	@Success      200  {object}  models.FirewallStatus
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /firewall/status [get]
func (s *FirewallService) GetStatus(c echo.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.conn == nil || s.table == nil {
		return c.JSON(http.StatusOK, models.FirewallStatus{
			Enabled:   false,
			RuleCount: 0,
			TableName: tableName,
		})
	}

	ruleCount := 0
	for _, chain := range s.chains {
		rules, err := s.conn.GetRules(s.table, chain)
		if err == nil {
			ruleCount += len(rules)
		}
	}

	return c.JSON(http.StatusOK, models.FirewallStatus{
		Enabled:   true,
		RuleCount: ruleCount,
		TableName: tableName,
	})
}

// ListRules returns all firewall rules
//
//	@Summary      list firewall rules
//	@Description  returns all rules in the filter table
//	@Tags         firewall
//	@Produce      json
//	@Success      200  {array}   models.FirewallRule
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /firewall/rules [get]
func (s *FirewallService) ListRules(c echo.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.conn == nil || s.table == nil {
		return c.JSON(http.StatusOK, []models.FirewallRule{})
	}

	var allRules []models.FirewallRule

	for chainName, chain := range s.chains {
		rules, err := s.conn.GetRules(s.table, chain)
		if err != nil {
			s.Logger.Error("failed to get rules", "chain", chainName, "error", err)
			continue
		}

		for _, rule := range rules {
			parsed := s.parseRule(rule, chainName)
			if parsed != nil {
				allRules = append(allRules, *parsed)
			}
		}
	}

	return c.JSON(http.StatusOK, allRules)
}

// parseRule converts an nftables rule to our FirewallRule model
func (s *FirewallService) parseRule(rule *nftables.Rule, chainName string) *models.FirewallRule {
	// Skip safety rules (system-managed) - they shouldn't be visible to users
	if s.isSafetyRule(rule) {
		return nil
	}

	fr := &models.FirewallRule{
		Handle: rule.Handle,
		Chain:  chainName,
	}

	for _, e := range rule.Exprs {
		switch v := e.(type) {
		case *expr.Verdict:
			switch v.Kind {
			case expr.VerdictAccept:
				fr.Action = "accept"
			case expr.VerdictDrop:
				fr.Action = "drop"
			}
		case *expr.Cmp:
			// Try to extract port or protocol info
			if len(v.Data) == 1 {
				// Protocol
				switch v.Data[0] {
				case unix.IPPROTO_TCP:
					fr.Protocol = "tcp"
				case unix.IPPROTO_UDP:
					fr.Protocol = "udp"
				}
			} else if len(v.Data) == 2 {
				// Port (big endian)
				fr.Port = binaryutil.BigEndian.Uint16(v.Data)
			} else if len(v.Data) == 4 {
				// IP address
				ip := net.IP(v.Data)
				fr.SourceIP = ip.String()
			}
		}
	}

	// Only return rules with an action (these are our managed rules)
	if fr.Action == "" {
		return nil
	}

	return fr
}

// AddRule creates a new firewall rule
//
//	@Summary      add firewall rule
//	@Description  creates a new rule in the specified chain
//	@Tags         firewall
//	@Accept       json
//	@Produce      json
//	@Param        rule  body      models.CreateRuleRequest  true  "Rule configuration"
//	@Success      201   {object}  models.FirewallRule
//	@Failure      400   {object}  models.HTTPError
//	@Failure      401   {object}  models.HTTPError
//	@Failure      403   {object}  models.HTTPError
//	@Failure      500   {object}  models.HTTPError
//	@Router       /firewall/rules [post]
func (s *FirewallService) AddRule(c echo.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var req models.CreateRuleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate chain
	chain, ok := s.chains[req.Chain]
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid chain: must be input, forward, or output")
	}

	// Validate protocol is required when port is specified
	if req.Port > 0 && req.Protocol == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "protocol (tcp or udp) is required when specifying a port")
	}

	// Build rule expressions
	exprs := s.buildRuleExprs(req)

	rule := &nftables.Rule{
		Table: s.table,
		Chain: chain,
		Exprs: exprs,
	}

	s.conn.AddRule(rule)

	if err := s.conn.Flush(); err != nil {
		return s.Dispatcher.NewInternalServerError("failed to add rule", err)
	}

	// Get the rule back to get its handle
	rules, err := s.conn.GetRules(s.table, chain)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("failed to get rules after add", err)
	}

	var createdRule *models.FirewallRule
	if len(rules) > 0 {
		lastRule := rules[len(rules)-1]
		createdRule = &models.FirewallRule{
			Handle:   lastRule.Handle,
			Chain:    req.Chain,
			Protocol: req.Protocol,
			Port:     req.Port,
			SourceIP: req.SourceIP,
			Action:   req.Action,
			Comment:  req.Comment,
		}
	}

	// Persist rules
	if err := s.persistRules(); err != nil {
		s.Logger.Error("failed to persist rules", "error", err)
	}

	return c.JSON(http.StatusCreated, createdRule)
}

// buildRuleExprs creates nftables expressions from a CreateRuleRequest
func (s *FirewallService) buildRuleExprs(req models.CreateRuleRequest) []expr.Any {
	var exprs []expr.Any

	// Add protocol match if specified
	if req.Protocol != "" {
		var proto byte
		switch req.Protocol {
		case "tcp":
			proto = unix.IPPROTO_TCP
		case "udp":
			proto = unix.IPPROTO_UDP
		}

		exprs = append(exprs,
			&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{proto},
			},
		)
	}

	// Add source IP match if specified
	if req.SourceIP != "" {
		ip := net.ParseIP(req.SourceIP)
		if ip != nil {
			ip4 := ip.To4()
			if ip4 != nil {
				exprs = append(exprs,
					&expr.Payload{
						DestRegister: 1,
						Base:         expr.PayloadBaseNetworkHeader,
						Offset:       12, // Source IP offset in IPv4 header
						Len:          4,
					},
					&expr.Cmp{
						Op:       expr.CmpOpEq,
						Register: 1,
						Data:     ip4,
					},
				)
			}
		}
	}

	// Add port match if specified
	if req.Port > 0 {
		exprs = append(exprs,
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseTransportHeader,
				Offset:       2, // Destination port offset
				Len:          2,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     binaryutil.BigEndian.PutUint16(req.Port),
			},
		)
	}

	// Add counter
	exprs = append(exprs, &expr.Counter{})

	// Add verdict
	var verdict expr.VerdictKind
	switch req.Action {
	case "accept":
		verdict = expr.VerdictAccept
	case "drop":
		verdict = expr.VerdictDrop
	}
	exprs = append(exprs, &expr.Verdict{Kind: verdict})

	return exprs
}

// DeleteRule removes a firewall rule by handle
//
//	@Summary      delete firewall rule
//	@Description  removes a rule by its handle
//	@Tags         firewall
//	@Param        handle  path  int  true  "Rule handle"
//	@Success      204
//	@Failure      400  {object}  models.HTTPError
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      404  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /firewall/rules/{handle} [delete]
func (s *FirewallService) DeleteRule(c echo.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	handleStr := c.Param("handle")
	handle, err := strconv.ParseUint(handleStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid handle")
	}

	// Find the rule
	var foundRule *nftables.Rule
	var foundChain *nftables.Chain

	for _, chain := range s.chains {
		rules, err := s.conn.GetRules(s.table, chain)
		if err != nil {
			continue
		}

		for _, rule := range rules {
			if rule.Handle == handle {
				foundRule = rule
				foundChain = chain
				break
			}
		}
		if foundRule != nil {
			break
		}
	}

	if foundRule == nil {
		return echo.NewHTTPError(http.StatusNotFound, "rule not found")
	}

	// Prevent deletion of safety rules
	if s.isSafetyRule(foundRule) {
		return echo.NewHTTPError(http.StatusForbidden, "cannot delete system safety rules (loopback/established connections)")
	}

	// Delete the rule
	s.conn.DelRule(&nftables.Rule{
		Table:  s.table,
		Chain:  foundChain,
		Handle: handle,
	})

	if err := s.conn.Flush(); err != nil {
		return s.Dispatcher.NewInternalServerError("failed to delete rule", err)
	}

	// Persist rules
	if err := s.persistRules(); err != nil {
		s.Logger.Error("failed to persist rules", "error", err)
	}

	return c.NoContent(http.StatusNoContent)
}

// ReorderRules changes the order of rules in a chain
//
//	@Summary      reorder firewall rules
//	@Description  reorders rules in a chain by re-adding them in the specified order
//	@Tags         firewall
//	@Accept       json
//	@Produce      json
//	@Param        order  body      models.ReorderRulesRequest  true  "New rule order"
//	@Success      200    {array}   models.FirewallRule
//	@Failure      400    {object}  models.HTTPError
//	@Failure      401    {object}  models.HTTPError
//	@Failure      403    {object}  models.HTTPError
//	@Failure      500    {object}  models.HTTPError
//	@Router       /firewall/rules/reorder [post]
func (s *FirewallService) ReorderRules(c echo.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var req models.ReorderRulesRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	chain, ok := s.chains[req.Chain]
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid chain: must be input, forward, or output")
	}

	// Get current rules in the chain
	currentRules, err := s.conn.GetRules(s.table, chain)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("failed to get rules", err)
	}

	// Build a map of handle -> rule for quick lookup
	ruleMap := make(map[uint64]*nftables.Rule)
	for _, rule := range currentRules {
		ruleMap[rule.Handle] = rule
	}

	// Validate all handles exist
	for _, handle := range req.Handles {
		if _, exists := ruleMap[handle]; !exists {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("rule with handle %d not found in chain %s", handle, req.Chain))
		}
	}

	// Delete all rules in the chain
	for _, rule := range currentRules {
		s.conn.DelRule(rule)
	}

	if err := s.conn.Flush(); err != nil {
		return s.Dispatcher.NewInternalServerError("failed to delete rules for reorder", err)
	}

	// Re-add rules in the new order
	for _, handle := range req.Handles {
		oldRule := ruleMap[handle]
		s.conn.AddRule(&nftables.Rule{
			Table: s.table,
			Chain: chain,
			Exprs: oldRule.Exprs,
		})
	}

	if err := s.conn.Flush(); err != nil {
		return s.Dispatcher.NewInternalServerError("failed to re-add rules", err)
	}

	// Persist the new order
	if err := s.persistRules(); err != nil {
		s.Logger.Error("failed to persist rules", "error", err)
	}

	// Return updated rules
	newRules, err := s.conn.GetRules(s.table, chain)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("failed to get reordered rules", err)
	}

	var result []models.FirewallRule
	for _, rule := range newRules {
		parsed := s.parseRule(rule, req.Chain)
		if parsed != nil {
			result = append(result, *parsed)
		}
	}

	return c.JSON(http.StatusOK, result)
}

// persistRules saves the current rules to a JSON file
func (s *FirewallService) persistRules() error {
	var allRules []models.FirewallRule

	for chainName, chain := range s.chains {
		rules, err := s.conn.GetRules(s.table, chain)
		if err != nil {
			continue
		}

		for _, rule := range rules {
			parsed := s.parseRule(rule, chainName)
			if parsed != nil {
				allRules = append(allRules, *parsed)
			}
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(s.dataDir, 0o755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	filePath := filepath.Join(s.dataDir, persistenceFile)
	data, err := json.MarshalIndent(allRules, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write rules file: %w", err)
	}

	s.Logger.Info("persisted firewall rules", "count", len(allRules), "file", filePath)

	// Call the backup callback if set (for creating firewall backups)
	if s.onRulesChanged != nil {
		go s.onRulesChanged(allRules)
	}

	return nil
}

// loadPersistedRules loads rules from the JSON file and applies them
func (s *FirewallService) loadPersistedRules() error {
	filePath := filepath.Join(s.dataDir, persistenceFile)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No persisted rules, that's fine
		}
		return fmt.Errorf("failed to read rules file: %w", err)
	}

	var rules []models.FirewallRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return fmt.Errorf("failed to unmarshal rules: %w", err)
	}

	// Clear existing rules in our chains to prevent duplicates on restart
	for _, chain := range s.chains {
		existingRules, err := s.conn.GetRules(s.table, chain)
		if err != nil {
			continue
		}
		for _, rule := range existingRules {
			s.conn.DelRule(rule)
		}
	}
	if err := s.conn.Flush(); err != nil {
		s.Logger.Warn("failed to clear existing rules", "error", err)
	}

	// Apply each rule
	for _, rule := range rules {
		chain, ok := s.chains[rule.Chain]
		if !ok {
			s.Logger.Warn("skipping rule with unknown chain", "chain", rule.Chain)
			continue
		}

		req := models.CreateRuleRequest{
			Chain:    rule.Chain,
			Protocol: rule.Protocol,
			Port:     rule.Port,
			SourceIP: rule.SourceIP,
			Action:   rule.Action,
			Comment:  rule.Comment,
		}

		exprs := s.buildRuleExprs(req)

		s.conn.AddRule(&nftables.Rule{
			Table: s.table,
			Chain: chain,
			Exprs: exprs,
		})
	}

	if err := s.conn.Flush(); err != nil {
		return fmt.Errorf("failed to apply persisted rules: %w", err)
	}

	s.Logger.Info("loaded persisted firewall rules", "count", len(rules))
	return nil
}
