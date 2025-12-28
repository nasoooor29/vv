# Visory

A comprehensive server management dashboard for managing Docker containers, QEMU/KVM virtual machines, system storage, and more.

## Features

- Docker container management
- QEMU/KVM virtual machine management with VNC access
- System storage and ISO management
- User authentication with OAuth (Google, GitHub)
- Role-based access control (RBAC)
- System monitoring and logs
- Discord notifications

## Quick Start (Using Pre-built Binaries)

### 1. Download the Binary

Download the latest release for your platform from the [GitHub Releases](https://github.com/nasoooor29/visory/releases) page.

Available platforms:
- **Linux** (amd64, arm64)
- **macOS** (Universal binary - amd64/arm64)
- **Windows** (amd64, arm64)

```bash
# Example for Linux amd64
curl -LO https://github.com/nasoooor29/visory/releases/latest/download/visory_Linux_x86_64.tar.gz
tar -xzf visory_Linux_x86_64.tar.gz

# Example for Linux arm64
curl -LO https://github.com/nasoooor29/visory/releases/latest/download/visory_Linux_arm64.tar.gz
tar -xzf visory_Linux_arm64.tar.gz

# Example for macOS (Universal)
curl -LO https://github.com/nasoooor29/visory/releases/latest/download/visory_Darwin_all.tar.gz
tar -xzf visory_Darwin_all.tar.gz

# Example for Windows (use PowerShell)
Invoke-WebRequest -Uri https://github.com/nasoooor29/visory/releases/latest/download/visory_Windows_x86_64.zip -OutFile visory.zip
Expand-Archive visory.zip -DestinationPath .
```

### 2. Configure Environment Variables

Create a `.env` file in the same directory as the binary:

```bash
# Server Configuration
PORT=9999
APP_ENV=production

# Database (SQLite file path)
BLUEPRINT_DB_DATABASE=visory.db

# Session Secret (REQUIRED - generate a secure random string)
# Generate with: openssl rand -base64 32
SESSION_SECRET=your-secure-random-secret-here

# OAuth Configuration (optional)
# Google OAuth - https://console.cloud.google.com/
GOOGLE_OAUTH_KEY=your-google-client-id
GOOGLE_OAUTH_SECRET=your-google-client-secret

# GitHub OAuth - https://github.com/settings/developers
GITHUB_OAUTH_KEY=your-github-client-id
GITHUB_OAUTH_SECRET=your-github-client-secret

# OAuth Callback URL (update with your domain)
OAUTH_CALLBACK_URL=http://localhost:9999/api/auth/oauth/callback

# Discord Notifications (optional)
DISCORD_WEBHOOK_URL=your-discord-webhook-url
DISCORD_NOTIFY_ON_ERROR=true
DISCORD_NOTIFY_ON_WARN=false
DISCORD_NOTIFY_ON_INFO=false
```

### 3. Run the Application

```bash
# Linux/macOS
./visory

# Windows
.\visory.exe
```

### 4. Access the Dashboard

Open your browser and navigate to:
- **Dashboard**: http://localhost:9999
- **API Documentation**: http://localhost:9999/api/docs/swagger

## System Requirements

### For Docker Management
- Docker Engine installed and running
- User must have permissions to access Docker socket

### For QEMU/KVM Virtual Machines
- libvirt installed and running
- QEMU/KVM packages installed
- User must be in the `libvirt` group

```bash
# Ubuntu/Debian
sudo apt install qemu-kvm libvirt-daemon-system libvirt-clients
sudo usermod -aG libvirt $USER

# Fedora/RHEL
sudo dnf install qemu-kvm libvirt
sudo usermod -aG libvirt $USER
```

## Running as a Service (Linux)

### Using systemd

Create a systemd service file at `/etc/systemd/system/visory.service`:

```ini
[Unit]
Description=Visory Server Management Dashboard
After=network.target docker.service libvirtd.service

[Service]
Type=simple
User=your-username
WorkingDirectory=/path/to/visory
ExecStart=/path/to/visory/visory
Restart=always
RestartSec=5
EnvironmentFile=/path/to/visory/.env

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable visory
sudo systemctl start visory

# Check status
sudo systemctl status visory

# View logs
sudo journalctl -u visory -f
```

<!-- ## Running with Docker -->
<!---->
<!-- You can also run Visory using Docker: -->
<!---->
<!-- ```bash -->
<!-- docker run -d \ -->
<!--   --name visory \ -->
<!--   -p 9999:9999 \ -->
<!--   -v /var/run/docker.sock:/var/run/docker.sock \ -->
<!--   -v visory-data:/data \ -->
<!--   -e PORT=9999 \ -->
<!--   -e APP_ENV=production \ -->
<!--   -e BLUEPRINT_DB_DATABASE=/data/visory.db \ -->
<!--   -e SESSION_SECRET=your-secure-secret \ -->
<!--   ghcr.io/nasoooor29/visory:latest -->
<!-- ``` -->
<!---->
<!-- Or using docker-compose: -->
<!---->
<!-- ```yaml -->
<!-- version: '3.8' -->
<!-- services: -->
<!--   visory: -->
<!--     image: ghcr.io/nasoooor29/visory:latest -->
<!--     ports: -->
<!--       - "9999:9999" -->
<!--     volumes: -->
<!--       - /var/run/docker.sock:/var/run/docker.sock -->
<!--       - visory-data:/data -->
<!--     environment: -->
<!--       - PORT=9999 -->
<!--       - APP_ENV=production -->
<!--       - BLUEPRINT_DB_DATABASE=/data/visory.db -->
<!--       - SESSION_SECRET=your-secure-secret -->
<!--     restart: unless-stopped -->
<!---->
<!-- volumes: -->
<!--   visory-data: -->
<!-- ``` -->

## First-Time Setup

1. **Create Initial User**: The first user to register (via the registration form or OAuth) is automatically granted full admin privileges (`user_admin` role). This ensures you have immediate access to all features without manual configuration.
2. **Configure OAuth** (optional): Set up Google or GitHub OAuth for easier authentication. See [OAuth Setup](OAUTH_SETUP.md)
3. **Set Up RBAC**: Configure role-based permissions for additional users. Subsequent users will be assigned the default `user` role and will need an admin to grant them additional permissions.

## Available RBAC Permissions

| Permission | Description |
|------------|-------------|
| `user_admin` | Full admin access (bypasses all checks) |
| `docker_read` | View Docker containers |
| `docker_write` | Create Docker containers |
| `docker_update` | Modify Docker containers |
| `docker_delete` | Delete Docker containers |
| `qemu_read` | View virtual machines |
| `qemu_write` | Create virtual machines |
| `qemu_update` | Modify virtual machines |
| `qemu_delete` | Delete virtual machines |
| `event_viewer` | View system events |
| `event_manager` | Manage system events |
| `settings_manager` | Manage system settings |
| `audit_log_viewer` | View audit logs |
| `health_checker` | Access health monitoring |

## Development

For development setup, see the [Getting Started](content/docs/getting-started.mdx) documentation.

### Quick Development Setup

```bash
# Clone the repository
git clone https://github.com/nasoooor29/visory.git
cd visory

# Install dependencies
go mod download
cd frontend && bun install && cd ..

# Run with hot reload
make tmux
```

## MakeFile Commands

| Command | Description |
|---------|-------------|
| `make all` | Build with tests |
| `make build` | Build the application |
| `make run` | Run the application |
| `make test` | Run the test suite |
| `make watch` | Live reload the application |
| `make clean` | Clean up build artifacts |

## Documentation

- [Authentication](content/docs/authentication.mdx)
- [OAuth Setup](OAUTH_SETUP.md)
- [RBAC Configuration](content/docs/rbac.mdx)
- [Docker Management](content/docs/docker.mdx)
- [Virtual Machines](content/docs/virtual-machines.mdx)
- [Storage Management](content/docs/storage.mdx)
- [Logs & Monitoring](content/docs/logs.mdx)

## License

This project is licensed under the MIT License.
