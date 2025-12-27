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
)

type FirewallService struct {
	Dispatcher *utils.Dispatcher
	Logger     *slog.Logger
	mu         sync.RWMutex
	conn       *nftables.Conn
	table      *nftables.Table
	chains     map[string]*nftables.Chain
	dataDir    string
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
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	filePath := filepath.Join(s.dataDir, persistenceFile)
	data, err := json.MarshalIndent(allRules, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write rules file: %w", err)
	}

	s.Logger.Info("persisted firewall rules", "count", len(allRules), "file", filePath)
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
