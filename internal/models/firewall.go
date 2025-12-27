package models

// FirewallRule represents a single nftables rule
type FirewallRule struct {
	Handle   uint64 `json:"handle"`    // nftables rule handle (unique ID)
	Chain    string `json:"chain"`     // "input", "forward", or "output"
	Protocol string `json:"protocol"`  // "tcp", "udp", or "" for any
	Port     uint16 `json:"port"`      // destination port (0 = any)
	SourceIP string `json:"source_ip"` // source IP/CIDR or "" for any
	Action   string `json:"action"`    // "accept" or "drop"
	Comment  string `json:"comment"`   // optional user comment
}

// FirewallStatus represents the current firewall state
type FirewallStatus struct {
	Enabled   bool   `json:"enabled"`
	RuleCount int    `json:"rule_count"`
	TableName string `json:"table_name"`
}

// CreateRuleRequest is the request body for creating a new firewall rule
type CreateRuleRequest struct {
	Chain    string `json:"chain" validate:"required,oneof=input forward output"`
	Protocol string `json:"protocol" validate:"omitempty,oneof=tcp udp"`
	Port     uint16 `json:"port"`
	SourceIP string `json:"source_ip"`
	Action   string `json:"action" validate:"required,oneof=accept drop"`
	Comment  string `json:"comment"`
}
