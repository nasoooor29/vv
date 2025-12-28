package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"visory/pkg/depscan"
)

type OutputFormat string

const (
	FormatText     OutputFormat = "text"
	FormatJSON     OutputFormat = "json"
	FormatMarkdown OutputFormat = "markdown"
)

type PolicyConfig struct {
	AllowedLicenses []string
	DeniedLicenses  []string
	FailOnDenied    bool
}

func main() {
	format := flag.String("format", "text", "Output format: text, json, markdown")
	jsonFlag := flag.Bool("json", false, "Output as JSON (shorthand for --format json)")
	markdownFlag := flag.Bool("markdown", false, "Output as Markdown (shorthand for --format markdown)")
	failOnLicenses := flag.String("fail-on", "", "Comma-separated list of licenses to fail on (e.g., GPL,AGPL)")
	onlyIndirect := flag.Bool("indirect-only", false, "Only show indirect dependencies")
	onlyDirect := flag.Bool("direct-only", false, "Only show direct dependencies")
	allowList := flag.String("allow", "", "Comma-separated list of allowed licenses")
	denyList := flag.String("deny", "", "Comma-separated list of denied licenses")
	goModPath := flag.String("go-mod", "", "Path to go.mod file (defaults to ./go.mod)")
	sortBy := flag.String("sort", "path", "Sort by: path, license, or version")
	outputFile := flag.String("out", "", "Write output to file instead of stdout")
	summary := flag.Bool("summary", false, "Show summary statistics only")
	warnNonMIT := flag.Bool("warn-non-mit", false, "Send Discord warning for non-MIT licenses (requires DISCORD_WEBHOOK_URL env var)")

	flag.Parse()

	// Get Discord webhook URL from environment variable
	discordWebhook := os.Getenv("DISCORD_WEBHOOK_URL")

	// Handle shorthand flags
	if *jsonFlag {
		*format = "json"
	}
	if *markdownFlag {
		*format = "markdown"
	}

	// Determine go.mod path
	if *goModPath == "" {
		*goModPath = "./go.mod"
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(*goModPath)
	if err != nil {
		fatal("failed to resolve go.mod path: %v", err)
	}

	// Collect dependencies
	deps, err := depscan.CollectDependencies(absPath)
	if err != nil {
		fatal("failed to collect dependencies: %v", err)
	}

	// Filter by policy
	filtered := filterDependencies(deps, *onlyDirect, *onlyIndirect, *allowList, *denyList)

	// Sort dependencies
	sortDependencies(filtered, *sortBy)

	// Check for policy violations
	if *failOnLicenses != "" {
		violatingDeps := checkFailOnLicenses(filtered, *failOnLicenses)
		if len(violatingDeps) > 0 {
			if *format == "text" || *format == "" {
				fmt.Fprintf(os.Stderr, "ERROR: Found dependencies with prohibited licenses:\n")
				for _, dep := range violatingDeps {
					fmt.Fprintf(os.Stderr, "  %s (%s) %s\n", dep.Path, dep.Version, dep.License)
				}
			}
			os.Exit(1)
		}
	}

	// Check for non-MIT licenses and send Discord warning
	if *warnNonMIT {
		if discordWebhook == "" {
			fatal("DISCORD_WEBHOOK_URL environment variable is required when using --warn-non-mit")
		}

		nonMITDeps := findNonMITLicenses(filtered)
		if len(nonMITDeps) > 0 {
			fmt.Fprintf(os.Stderr, "Found %d non-MIT dependencies, sending Discord notification...\n", len(nonMITDeps))
			if err := sendDiscordWarning(discordWebhook, nonMITDeps); err != nil {
				fatal("failed to send Discord notification: %v", err)
			}
			fmt.Fprintf(os.Stderr, "Discord notification sent successfully for %d non-MIT dependencies\n", len(nonMITDeps))
		} else {
			fmt.Fprintf(os.Stderr, "All dependencies have MIT-compatible licenses\n")
		}
	}

	// Format output
	var output string
	switch OutputFormat(*format) {
	case FormatJSON:
		output = formatJSON(filtered, *summary)
	case FormatMarkdown:
		output = formatMarkdown(filtered, *summary)
	default:
		output = formatText(filtered, *summary)
	}

	// Write output
	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, []byte(output), 0644); err != nil {
			fatal("failed to write output file: %v", err)
		}
		fmt.Printf("Output written to %s\n", *outputFile)
	} else {
		fmt.Print(output)
	}
}

func filterDependencies(deps []depscan.Dependency, directOnly, indirectOnly bool, allowList, denyList string) []depscan.Dependency {
	var filtered []depscan.Dependency

	allowedLicenses := parseCSV(allowList)
	deniedLicenses := parseCSV(denyList)

	for _, dep := range deps {
		// Apply direct/indirect filter
		if directOnly && dep.Indirect {
			continue
		}
		if indirectOnly && !dep.Indirect {
			continue
		}

		// Apply allow list (if specified, only include allowed)
		if len(allowedLicenses) > 0 && !stringInSlice(dep.License, allowedLicenses) {
			continue
		}

		// Apply deny list (exclude denied)
		if len(deniedLicenses) > 0 && stringInSlice(dep.License, deniedLicenses) {
			continue
		}

		filtered = append(filtered, dep)
	}

	return filtered
}

func checkFailOnLicenses(deps []depscan.Dependency, failOnList string) []depscan.Dependency {
	failLicenses := parseCSV(failOnList)
	var violating []depscan.Dependency

	for _, dep := range deps {
		if stringInSlice(dep.License, failLicenses) {
			violating = append(violating, dep)
		}
	}

	return violating
}

func sortDependencies(deps []depscan.Dependency, sortBy string) {
	// Simple bubble sort for demonstration
	switch sortBy {
	case "license":
		for i := 0; i < len(deps); i++ {
			for j := i + 1; j < len(deps); j++ {
				if deps[i].License > deps[j].License {
					deps[i], deps[j] = deps[j], deps[i]
				}
			}
		}
	case "version":
		for i := 0; i < len(deps); i++ {
			for j := i + 1; j < len(deps); j++ {
				if deps[i].Version > deps[j].Version {
					deps[i], deps[j] = deps[j], deps[i]
				}
			}
		}
	default: // path (default)
		for i := 0; i < len(deps); i++ {
			for j := i + 1; j < len(deps); j++ {
				if deps[i].Path > deps[j].Path {
					deps[i], deps[j] = deps[j], deps[i]
				}
			}
		}
	}
}

func formatText(deps []depscan.Dependency, summaryOnly bool) string {
	var sb strings.Builder

	if summaryOnly {
		stats := calculateStats(deps)
		sb.WriteString(fmt.Sprintf("Total Dependencies: %d\n", len(deps)))
		sb.WriteString(fmt.Sprintf("Direct Dependencies: %d\n", stats.Direct))
		sb.WriteString(fmt.Sprintf("Indirect Dependencies: %d\n", stats.Indirect))
		sb.WriteString(fmt.Sprintf("Licenses Found: %d\n", len(stats.Licenses)))
		sb.WriteString("\nLicense Distribution:\n")
		for license, count := range stats.Licenses {
			sb.WriteString(fmt.Sprintf("  %s: %d\n", license, count))
		}
		return sb.String()
	}

	sb.WriteString("DEPENDENCY REPORT\n")
	sb.WriteString("=================\n\n")

	for _, dep := range deps {
		directStr := "indirect"
		if !dep.Indirect {
			directStr = "direct"
		}

		sb.WriteString(fmt.Sprintf("Name: %s\n", dep.Path))
		sb.WriteString(fmt.Sprintf("Version: %s\n", dep.Version))
		sb.WriteString(fmt.Sprintf("Type: %s\n", directStr))
		sb.WriteString(fmt.Sprintf("License: %s\n", dep.License))
		if dep.LicenseFile != "" {
			sb.WriteString(fmt.Sprintf("License File: %s\n", dep.LicenseFile))
		}
		sb.WriteString(fmt.Sprintf("Directory: %s\n", dep.Dir))
		sb.WriteString("\n")
	}

	stats := calculateStats(deps)
	sb.WriteString("SUMMARY\n")
	sb.WriteString("=======\n")
	sb.WriteString(fmt.Sprintf("Total Dependencies: %d\n", len(deps)))
	sb.WriteString(fmt.Sprintf("Direct: %d | Indirect: %d\n", stats.Direct, stats.Indirect))
	sb.WriteString(fmt.Sprintf("Unique Licenses: %d\n", len(stats.Licenses)))
	sb.WriteString("\nLicense Distribution:\n")
	for license, count := range stats.Licenses {
		sb.WriteString(fmt.Sprintf("  %s: %d\n", license, count))
	}

	return sb.String()
}

func formatJSON(deps []depscan.Dependency, summaryOnly bool) string {
	if summaryOnly {
		stats := calculateStats(deps)
		output := map[string]interface{}{
			"total":    len(deps),
			"direct":   stats.Direct,
			"indirect": stats.Indirect,
			"licenses": stats.Licenses,
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		return string(data) + "\n"
	}

	data, _ := json.MarshalIndent(deps, "", "  ")
	return string(data) + "\n"
}

func formatMarkdown(deps []depscan.Dependency, summaryOnly bool) string {
	var sb strings.Builder

	sb.WriteString("# Dependency Report\n\n")

	if summaryOnly {
		stats := calculateStats(deps)
		sb.WriteString("## Summary\n\n")
		sb.WriteString(fmt.Sprintf("- **Total Dependencies:** %d\n", len(deps)))
		sb.WriteString(fmt.Sprintf("- **Direct:** %d\n", stats.Direct))
		sb.WriteString(fmt.Sprintf("- **Indirect:** %d\n", stats.Indirect))
		sb.WriteString(fmt.Sprintf("- **Unique Licenses:** %d\n\n", len(stats.Licenses)))

		sb.WriteString("## License Distribution\n\n")
		sb.WriteString("| License | Count |\n")
		sb.WriteString("|---------|-------|\n")
		for license, count := range stats.Licenses {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", license, count))
		}
		return sb.String()
	}

	sb.WriteString("## Dependencies\n\n")
	sb.WriteString("| Module | Version | Type | License | License File |\n")
	sb.WriteString("|--------|---------|------|---------|---------------|\n")

	for _, dep := range deps {
		depType := "Indirect"
		if !dep.Indirect {
			depType = "Direct"
		}

		licenseFile := "-"
		if dep.LicenseFile != "" {
			licenseFile = dep.LicenseFile
		}

		sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s | %s |\n",
			dep.Path, dep.Version, depType, dep.License, licenseFile))
	}

	stats := calculateStats(deps)
	sb.WriteString("\n## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Dependencies:** %d\n", len(deps)))
	sb.WriteString(fmt.Sprintf("- **Direct:** %d\n", stats.Direct))
	sb.WriteString(fmt.Sprintf("- **Indirect:** %d\n", stats.Indirect))
	sb.WriteString(fmt.Sprintf("- **Unique Licenses:** %d\n\n", len(stats.Licenses)))

	sb.WriteString("## License Distribution\n\n")
	sb.WriteString("| License | Count |\n")
	sb.WriteString("|---------|-------|\n")
	for license, count := range stats.Licenses {
		sb.WriteString(fmt.Sprintf("| %s | %d |\n", license, count))
	}

	return sb.String()
}

type Stats struct {
	Direct   int
	Indirect int
	Licenses map[string]int
}

func calculateStats(deps []depscan.Dependency) Stats {
	stats := Stats{
		Licenses: make(map[string]int),
	}

	for _, dep := range deps {
		if dep.Indirect {
			stats.Indirect++
		} else {
			stats.Direct++
		}
		stats.Licenses[dep.License]++
	}

	return stats
}

func parseCSV(input string) []string {
	if input == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func stringInSlice(target string, slice []string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}

// findNonMITLicenses returns all dependencies that don't have an MIT license
func findNonMITLicenses(deps []depscan.Dependency) []depscan.Dependency {
	var nonMIT []depscan.Dependency
	for _, dep := range deps {
		license := strings.ToUpper(dep.License)
		if !strings.Contains(license, "MIT") {
			nonMIT = append(nonMIT, dep)
		}
	}
	return nonMIT
}

// sendDiscordWarning sends a warning notification to Discord about non-MIT licenses
func sendDiscordWarning(webhookURL string, deps []depscan.Dependency) error {
	// Build the message
	var depList strings.Builder
	for i, dep := range deps {
		if i >= 10 {
			depList.WriteString(fmt.Sprintf("\n... and %d more", len(deps)-10))
			break
		}
		depList.WriteString(fmt.Sprintf("• `%s` (%s) - **%s**\n", dep.Path, dep.Version, dep.License))
	}

	// Group licenses by type
	licenseCount := make(map[string]int)
	for _, dep := range deps {
		licenseCount[dep.License]++
	}

	var licenseSummary strings.Builder
	for license, count := range licenseCount {
		licenseSummary.WriteString(fmt.Sprintf("• %s: %d\n", license, count))
	}

	embed := discordEmbed{
		Title:       "[WARNING] Non-MIT Licenses Detected",
		Description: fmt.Sprintf("Found **%d** dependencies with non-MIT licenses that may require review.", len(deps)),
		Color:       0xFFA500, // Orange
		Fields: []discordEmbedField{
			{
				Name:   "License Summary",
				Value:  licenseSummary.String(),
				Inline: false,
			},
			{
				Name:   "Dependencies",
				Value:  depList.String(),
				Inline: false,
			},
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Footer: &discordFooter{
			Text: "Visory License Scanner",
		},
	}

	msg := discordWebhookMessage{
		Username: "Visory License Scanner",
		Embeds:   []discordEmbed{embed},
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal discord message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create discord webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send discord webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("discord webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// Discord webhook types
type discordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Fields      []discordEmbedField `json:"fields,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
	Footer      *discordFooter      `json:"footer,omitempty"`
}

type discordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type discordFooter struct {
	Text string `json:"text"`
}

type discordWebhookMessage struct {
	Username string         `json:"username,omitempty"`
	Embeds   []discordEmbed `json:"embeds,omitempty"`
}
