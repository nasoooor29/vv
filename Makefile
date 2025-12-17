# Simple Makefile for a Go project

# Build the application
# all: build test

build:
	@echo "Building..."
	@go build -o visory cmd/api/main.go

release:
	@if [ -z "$(VERSION)" ]; then echo "ERROR: VERSION is required. Usage: make release VERSION=v1.0.0"; exit 1; fi
	@echo "Building for all platforms..."
	@echo "Building for Mac"
	@GOOS=darwin GOARCH=amd64 go build -o bin/visory-darwin-amd64 cmd/api/main.go
	@GOOS=darwin GOARCH=arm64 go build -o bin/visory-darwin-arm64 cmd/api/main.go

	@echo "Building for Windows"
	@GOOS=windows GOARCH=amd64 go build -o bin/visory-windows-amd64.exe cmd/api/main.go
	@GOOS=windows GOARCH=arm64 go build -o bin/visory-windows-arm64.exe cmd/api/main.go

	@echo "Building for Linux"
	@GOOS=linux GOARCH=amd64 go build -o bin/visory-linux-amd64 cmd/api/main.go
	@GOOS=linux GOARCH=arm64 go build -o bin/visory-linux-arm64 cmd/api/main.go

	@echo "Building for FreeBSD"
	@GOOS=freebsd GOARCH=amd64 go build -o bin/visory-freebsd-amd64 cmd/api/main.go
	@GOOS=freebsd GOARCH=arm64 go build -o bin/visory-freebsd-arm64 cmd/api/main.go

	@echo "Building for OpenBSD"
	@GOOS=openbsd GOARCH=amd64 go build -o bin/visory-openbsd-amd64 cmd/api/main.go
	@GOOS=openbsd GOARCH=arm64 go build -o bin/visory-openbsd-arm64 cmd/api/main.go

	@echo "Generating checksums..."
	@cd bin && shasum -a 256 * > checksums.txt

	# @echo "Creating git tag $(VERSION)..."
	# @git tag -a $(VERSION) -m "Release $(VERSION)"
	# @git push origin $(VERSION)

	@echo "Publishing draft release on GitHub..."
	@gh release create $(VERSION) ./bin/* --draft --generate-notes

	@echo "All builds completed!"
	@echo "Binaries are located in the 'bin' directory."
	@echo "Draft release created: $(VERSION)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f visory

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Live Reload
back:
	@air

front:
	@cd ./frontend && bun run dev

tmux:
	@tmux split-window -h 'make front' \; select-pane -L \; send-keys 'make back' C-m

# .PHONY: all build release test watch front clean
