#!/bin/bash

# depscan Usage Examples
# ======================
# A collection of practical examples showing how to use the depscan library and CLI

## PART 1: CLI Examples
## ====================

# 1.1 Basic usage - show all dependencies
depscan

# 1.2 Summary only
depscan --summary

# 1.3 JSON output
depscan --json

# 1.4 Markdown report
depscan --markdown

# 1.5 Save to file
depscan --markdown --out LICENSES.md

## Filtering Examples

# 2.1 Only direct dependencies
depscan --direct-only

# 2.2 Only indirect dependencies
depscan --indirect-only

# 2.3 Only MIT licensed packages
depscan --allow "MIT"

# 2.4 Exclude GPL licenses
depscan --deny "GPL-2.0,GPL-3.0,AGPL-3.0"

# 2.5 Combine filters
depscan --direct-only --allow "MIT,Apache-2.0" --markdown

## Sorting Examples

# 3.1 Sort by license (default is path)
depscan --sort license

# 3.2 Sort by version
depscan --sort version

# 3.3 Sort by path explicitly
depscan --sort path

## Policy Enforcement

# 4.1 Fail if GPL is found (useful for CI/CD)
if depscan --fail-on "GPL-2.0,GPL-3.0,AGPL-3.0"; then
  echo "✓ No problematic licenses"
else
  echo "✗ Found prohibited licenses"
  exit 1
fi

# 4.2 Comprehensive policy check
depscan --fail-on "GPL,AGPL,SSPL" && \
  depscan --deny "UNKNOWN" --summary && \
  echo "✓ License policy check passed"

## Advanced Combinations

# 5.1 Generate audit report
depscan --markdown --sort license --out audit-$(date +%Y%m%d).md

# 5.2 Export direct dependencies as JSON
depscan --direct-only --json > direct-deps.json

# 5.3 Check and report unknown licenses
depscan --allow "UNKNOWN" --direct-only --sort path

# 5.4 Create allow-list for vendor compliance
depscan --json | jq '.[] | select(.License != "UNKNOWN") | .Path + ": " + .License'

# 5.5 Cross-module dependency audit
for module in "./service1" "./service2" "./service3"; do
  echo "=== $module ==="
  depscan --go-mod "$module/go.mod" --summary
done

## PART 2: Library Examples
## ========================

# See pkg/depscan/README.md for detailed library documentation

# 6.1 Basic Go library usage
go run - << 'EOF'
package main

import (
	"fmt"
	"log"
	"visory/pkg/depscan"
)

func main() {
	deps, err := depscan.CollectDependencies("./go.mod")
	if err != nil {
		log.Fatal(err)
	}
	
	for _, dep := range deps {
		fmt.Printf("%s@%s: %s\n", dep.Path, dep.Version, dep.License)
	}
}
EOF

# 6.2 Custom configuration
go run - << 'EOF'
package main

import (
	"fmt"
	"log"
	"visory/pkg/depscan"
)

func main() {
	cfg := depscan.Config{
		GoModPath: "./go.mod",
		LicenseNames: []string{"LICENSE", "COPYING", "COPYRIGHT"},
	}
	
	deps, err := depscan.CollectDependenciesWithConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Found %d dependencies\n", len(deps))
}
EOF

# 6.3 Filtering in code
go run - << 'EOF'
package main

import (
	"fmt"
	"log"
	"visory/pkg/depscan"
)

func main() {
	deps, err := depscan.CollectDependencies("./go.mod")
	if err != nil {
		log.Fatal(err)
	}
	
	// Filter for MIT licenses
	mit := depscan.FilterByLicense(deps, "MIT")
	fmt.Printf("MIT licenses: %d\n", len(mit))
	
	// Filter for GPL
	gpl := depscan.FilterByLicense(deps, "GPL-2.0", "GPL-3.0", "AGPL-3.0")
	if len(gpl) > 0 {
		fmt.Println("WARNING: GPL dependencies found!")
		for _, dep := range gpl {
			fmt.Printf("  - %s\n", dep.Path)
		}
	}
}
EOF

## PART 3: CI/CD Integration Examples
## ==================================

# 7.1 GitHub Actions
cat << 'EOF'
name: License Check
on: [push, pull_request]

jobs:
  license-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - name: Build depscan
        run: go build -o depscan ./cmd/depscan
      - name: Check for GPL licenses
        run: ./depscan --fail-on "GPL-2.0,GPL-3.0,AGPL-3.0"
      - name: Generate report
        if: always()
        run: ./depscan --markdown --out LICENSE_REPORT.md
      - name: Upload report
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: license-report
          path: LICENSE_REPORT.md
EOF

# 7.2 GitLab CI
cat << 'EOF'
license_check:
  stage: verify
  script:
    - go build -o depscan ./cmd/depscan
    - ./depscan --fail-on "GPL-2.0,GPL-3.0,AGPL-3.0"
    - ./depscan --markdown --out LICENSE_REPORT.md
  artifacts:
    paths:
      - LICENSE_REPORT.md
    expire_in: 30 days
EOF

# 7.3 Pre-commit hook
cat << 'EOF'
#!/bin/bash
# .git/hooks/pre-commit

go build -o /tmp/depscan ./cmd/depscan
if ! /tmp/depscan --fail-on "GPL-2.0,GPL-3.0,AGPL-3.0" > /dev/null 2>&1; then
  echo "ERROR: GPL licenses detected in dependencies"
  exit 1
fi
EOF

# 7.4 Makefile integration
cat << 'EOF'
.PHONY: license-check
license-check:
	@echo "Checking licenses..."
	@go build -o depscan ./cmd/depscan
	@./depscan --fail-on "GPL-2.0,GPL-3.0,AGPL-3.0"
	@echo "✓ License check passed"

.PHONY: license-report
license-report:
	@go build -o depscan ./cmd/depscan
	@./depscan --markdown --out LICENSES.md
	@echo "Report saved to LICENSES.md"
EOF

## PART 4: Compliance & Auditing
## =============================

# 8.1 Export all dependencies with licenses for audit
depscan --json | jq '.[] | {path: .Path, version: .Version, license: .License, indirect: .Indirect}'

# 8.2 Find all unknown licenses
depscan --json | jq '.[] | select(.License == "UNKNOWN") | .Path'

# 8.3 License distribution
depscan --json | jq 'group_by(.License) | map({license: .[0].License, count: length})'

# 8.4 Check license compliance for a specific SPDX list
APPROVED_LICENSES="MIT Apache-2.0 BSD-3-Clause ISC MPL-2.0"
for license in $APPROVED_LICENSES; do
  count=$(depscan --allow "$license" --json | jq 'length')
  echo "$license: $count"
done

# 8.5 Monitor license drift
echo "Generating baseline..."
depscan --json > /tmp/baseline.json

echo "Running tests..."
go test ./...

echo "Checking for changes..."
depscan --json > /tmp/current.json
diff <(jq 'sort_by(.Path)' /tmp/baseline.json) <(jq 'sort_by(.Path)' /tmp/current.json) || {
  echo "WARNING: Dependency licenses changed!"
  exit 1
}

## PART 5: Troubleshooting
## =======================

# 9.1 Verbose output (see each file being scanned)
depscan --json 2>&1 | head -50

# 9.2 Check specific module
depscan --json | jq '.[] | select(.Path | contains("github.com/example/lib"))'

# 9.3 Find all modules with missing license files
depscan --json | jq '.[] | select(.LicenseFile == "")'

# 9.4 List all unique licenses found
depscan --json | jq -r '.[] | .License' | sort -u

# 9.5 Generate compliance matrix
echo "Module,Version,Type,License" > compliance.csv
depscan --json | jq -r '.[] | [.Path, .Version, (if .Indirect then "Indirect" else "Direct" end), .License] | @csv' >> compliance.csv
