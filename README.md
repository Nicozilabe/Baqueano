

<p align="center">
    <img src="images/logo.png" alt="Baqueano Logo" width="120" />
</p>

# Baqueano - SSH Manager

[![Go](https://img.shields.io/badge/Go-1.27+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Release](https://img.shields.io/github/v/release/Nicozilabe/baqueano?style=for-the-badge)](https://github.com/Nicozilabe/baqueano/releases)
[![License](https://img.shields.io/github/license/Nicozilabe/baqueano?style=for-the-badge)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey?style=for-the-badge)](https://github.com/Nicozilabe/baqueano/releases)

> **This is a fork** of the original project by [Guillaume Archambault (@Gu1llaum-3)](https://github.com/Gu1llaum-3), maintained by [Nicozilabe](https://github.com/Nicozilabe). It adds optional Google Drive sync on top of the original project; the original MIT license and attribution are preserved.

> **A modern, interactive SSH Manager for your terminal**

> **Website:** [nicozilabe.github.io/baqueano](https://nicozilabe.github.io/baqueano/) · [Privacy Policy](https://nicozilabe.github.io/baqueano/privacy.html) · [Terms of Use](https://nicozilabe.github.io/baqueano/terms.html)

Baqueano is a beautiful command-line tool that transforms how you manage and connect to your SSH hosts. Built with Go and featuring an intuitive TUI interface, it makes SSH connection management effortless and enjoyable.

<p align="center">
    <a href="images/baq.gif" target="_blank">
        <img src="images/baq.gif" alt="Demo Baqueano Terminal" width="800" />
    </a>
    <br>
    <em>Click on the image to view in full size</em>
</p>

## Features

### **Core Capabilities**
- **Beautiful TUI Interface** - Navigate your SSH hosts with an elegant, interactive terminal UI
- **Quick Connect** - Connect to any host instantly through the TUI or the CLI with `baq <host>`
- **Port Forwarding** - Easy setup for Local, Remote, and Dynamic (SOCKS) forwarding with history persistence
- **Easy Management** - Add, edit, move, and manage SSH configurations seamlessly
- **Tag Support** - Organize your hosts with custom tags for better categorization; use the special `hidden` tag to exclude hosts from the list while keeping them connectable
- **Smart Search** - Find hosts quickly with built-in filtering and search
- **Connectivity Status** - Color-coded SSH connectivity indicators, checked in the background when the TUI starts and on demand
- **Smart Updates** - Automatic version checking with update notifications
- **Connection History** - Track your SSH connections with last login timestamps

### **Technical Features**
- **Secure** - Works directly with your existing `~/.ssh/config` file
- **Custom Config Support** - Use any SSH configuration file with the `-c` flag
- **SSH Include Support** - Full support for SSH Include directives to organize configurations across multiple files
- **SSH Options Support** - Add any SSH configuration option through intuitive forms
- **Automatic Conversion** - Seamlessly converts between command-line and config formats
- **Automatic Backups** - Backup configurations automatically before changes
- **Validation** - Prevent configuration errors with built-in validation
- **ProxyJump/ProxyCommand Support** - Secure connection tunneling through bastion hosts
- **Keyboard Shortcuts** - Power user navigation with vim-like shortcuts
- **Cross-platform** - Supports Linux, macOS (Intel & Apple Silicon), and Windows
- **Lightweight** - Single binary with no dependencies, zero configuration required

## Quick Start

### Installation

**Unix/Linux/macOS (One-line install):**
```bash
curl -sSL https://raw.githubusercontent.com/Nicozilabe/baqueano/main/install/unix.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/Nicozilabe/baqueano/main/install/windows.ps1 | iex
```

**Alternative methods:**

*Linux/macOS:*
```bash
# Download a specific release (asset shape: baqueano_<OS>_<arch>.tar.gz)
wget https://github.com/Nicozilabe/baqueano/releases/latest/download/baqueano_Linux_x86_64.tar.gz

# Extract and install
tar -xzf baqueano_Linux_x86_64.tar.gz
sudo mv baq /usr/local/bin/baq
```

*Windows:*
```powershell
# Download and extract
Invoke-WebRequest -Uri "https://github.com/Nicozilabe/baqueano/releases/latest/download/baqueano_Windows_x86_64.zip" -OutFile "baqueano_Windows_x86_64.zip"
Expand-Archive baqueano_Windows_x86_64.zip -DestinationPath C:\tools\
# Add C:\tools to your PATH environment variable
```

### Uninstalling

Removing Baqueano is just as easy as installing it.

**Unix/Linux/macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/Nicozilabe/baqueano/main/install/unix-uninstall.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/Nicozilabe/baqueano/main/install/windows-uninstall.ps1 | iex
```

> By default your configuration (history, backups, saved credentials) is kept.
> Add `PURGE_CONFIG=true` (Unix) or `-Purge` (Windows) to remove it too.

## Usage

### Interactive Mode

Launch Baqueano without arguments to enter the beautiful TUI interface:

```bash
baq
```

**Search (focused at startup):**
- Type to filter hosts by name, hostname or tag
- `↑/↓` or `Ctrl+P/Ctrl+N` - Select a host without leaving the search
- `Enter` - Connect to the selected host
- `Esc` - Clear the search (quits when the search is empty)
- `Tab` - Switch to the host table, where the shortcuts below work

**Host table:**
- `↑/↓` or `j/k` - Navigate hosts
- `Enter` - Connect to selected host
- `a` - Add new host
- `e` - Edit selected host
- `d` - Delete selected host
- `m` - Move host to another config file (requires SSH Include directives)
- `f` - Port forwarding setup
- `p` - Check the status of all hosts again
- `H` - Toggle hidden hosts visibility
- `q` - Quit
- `/` or `Tab` - Back to the search

**Status Indicators** (checked when the TUI starts, and again with `p`):
- **Online** - Host is reachable via SSH
- **Connecting** - Currently checking host connectivity
- **Offline** - Host is unreachable or SSH connection failed
- **Unknown** - Connectivity status not yet determined

**Sorting & Filtering:**
- `s` - Switch between sorting modes (name ↔ last login)
- `n` - Sort by **name** (alphabetical)
- `r` - Sort by **recent** (last login time)

The interactive forms will guide you through configuration:
- **Hostname/IP** - Server address
- **Username** - SSH user
- **Port** - SSH port (default: 22)
- **Identity File** - Private key path
- **ProxyJump** - Jump server for connection tunneling
- **ProxyCommand** - Jump command for connection tunneling
- **SSH Options** - Additional SSH options in `-o` format (e.g., `-o Compression=yes -o ServerAliveInterval=60`)
- **Tags** - Comma-separated tags for organization

### Port Forwarding

Baqueano provides an intuitive interface for setting up SSH port forwarding. Press `f` while selecting a host to open the port forwarding setup:

**Forward Types:**
- **Local (-L)** - Forward a local port to a remote host/port through the SSH connection
  - Example: Access a remote database on `localhost:5432` via local port `15432`
  - Use case: `ssh -L 15432:localhost:5432 server` → Database accessible on `localhost:15432`

- **Remote (-R)** - Forward a remote port back to a local host/port
  - Example: Expose local web server on remote host's port `8080`
  - Use case: `ssh -R 8080:localhost:3000 server` → Local app accessible from remote host's port 8080
  - **Requirements for external access:**
    - **SSH Server Config**: Add `GatewayPorts yes` to `/etc/ssh/sshd_config` and restart SSH service
    - **Firewall**: Open the remote port in the server's firewall (`ufw allow 8080` or equivalent)
    - **Port Availability**: Ensure the remote port is not already in use
    - **Bind Address**: Use `0.0.0.0` for external access, `127.0.0.1` for local-only

- **Dynamic (-D)** - Create a SOCKS proxy for secure browsing
  - Example: Route web traffic through the SSH connection
  - Use case: `ssh -D 1080 server` → Configure browser to use `localhost:1080` as SOCKS proxy
  - **Configuration requirements:**
    - **Browser Setup**: Configure SOCKS v5 proxy in browser settings
    - **DNS**: Enable "Proxy DNS when using SOCKS v5" for full privacy
    - **Applications**: Only SOCKS-aware applications will use the proxy
    - **Bind Address**: Use `127.0.0.1` for security (local access only)

**Port Forwarding Interface:**
- Choose forward type with ←/→ arrow keys
- Configure ports and addresses with guided forms
- Optional bind address configuration (defaults to 127.0.0.1)
- Real-time validation of port numbers and addresses
- **Port forwarding history** - Save frequently used configurations for quick reuse
- Connect automatically with configured forwarding options

**Troubleshooting Port Forwarding:**

*Remote Forwarding Issues:*
```bash
# Error: "remote port forwarding failed for listen port X"
# Solutions:
1. Check if port is already in use: ssh server "netstat -tln | grep :X"
2. Use a different port that's available
3. Enable GatewayPorts in SSH config for external access
```

*SSH Server Configuration for Remote Forwarding:*
```bash
# Edit SSH daemon config on the server:
sudo nano /etc/ssh/sshd_config

# Add or uncomment:
GatewayPorts yes

# Restart SSH service:
sudo systemctl restart sshd  # Ubuntu/Debian/CentOS 7+
# OR
sudo service ssh restart     # Older systems
```

*Firewall Configuration:*
```bash
# Ubuntu/Debian (UFW):
sudo ufw allow [port_number]

# CentOS/RHEL/Rocky (firewalld):
sudo firewall-cmd --add-port=[port_number]/tcp --permanent
sudo firewall-cmd --reload

# Check if port is accessible:
telnet [server_ip] [port_number]
```

*Dynamic Forwarding (SOCKS) Browser Setup:*
```
Firefox: about:preferences → Network Settings
- Manual proxy configuration
- SOCKS Host: localhost, Port: [your_port]
- SOCKS v5: ✓
- Proxy DNS when using SOCKS v5: ✓

Chrome: Launch with proxy
chrome --proxy-server="socks5://localhost:[your_port]"
```

### CLI Usage

Baqueano provides both command-line operations and an interactive TUI interface:

```bash
# Launch interactive TUI mode for browsing and connecting to hosts
baq

# Connect directly to a specific host (with history tracking)
baq my-server

# Execute a command on a remote host
baq my-server uptime

# Execute command with arguments
baq my-server ls -la /var/log

# Force TTY allocation for interactive commands
baq -t my-server sudo systemctl restart nginx

# Launch TUI with custom SSH config file
baq -c /path/to/custom/ssh_config

# Connect directly with custom SSH config file
baq my-server -c /path/to/custom/ssh_config

# Add a new host using interactive form
baq add

# Add a new host with pre-filled hostname
baq add hostname

# Add a new host with custom SSH config file
baq add hostname -c /path/to/custom/ssh_config

# Edit an existing host configuration
baq edit my-server

# Edit host with custom SSH config file
baq edit my-server -c /path/to/custom/ssh_config

# Move a host to another SSH config file (requires Include directives)
baq move my-server

# Move host with custom SSH config file (requires Include directives)
baq move my-server -c /path/to/custom/ssh_config

# Search for hosts (interactive filter)
baq search

# Print machine-readable info (JSON) for scripting
baq info prod-server
baq info prod-server --pretty

# With a custom SSH config file
baq -c /path/to/custom/ssh_config info prod-server

# Pipe to jq
baq info prod-server | jq -r '.result.target.hostname'
baq info prod-server | jq -r '.result.target.user'

# Show version information
baq --version

# Disable automatic update check (useful on air-gapped machines)
baq --no-update-check

# Show help and available commands
baq --help
```

### Host Info (JSON)

`baq info <hostname>` prints a single JSON object to stdout so you can script against it with `jq`.

```bash
# Extract fields
baq info prod-server | jq -r '.result.target.hostname'
baq info prod-server | jq -r '.result.target.port'

# Check not-found (exit code 2)
baq info does-not-exist | jq -r '.error.code'
```

### Shell Completion

Baqueano supports shell completion for host names, making it easy to connect to hosts without typing full names:

```bash
baq <TAB>           # Lists all available hosts
baq pro<TAB>        # Completes to hosts starting with "pro" (e.g., prod-server)
```

**Setup Instructions:**

**Bash:**
```bash
# Enable for current session
source <(baq completion bash)

# Enable permanently (add to ~/.bashrc)
echo 'source <(baq completion bash)' >> ~/.bashrc
```

**Zsh:**
```bash
# Enable for current session
source <(baq completion zsh)

# Enable permanently (add to ~/.zshrc)
echo 'source <(baq completion zsh)' >> ~/.zshrc
```

**Fish:**
```bash
# Enable for current session
baq completion fish | source

# Enable permanently
baq completion fish > ~/.config/fish/completions/baq.fish
```

**PowerShell:**
```powershell
# Enable for current session
baq completion powershell | Out-String | Invoke-Expression

# Enable permanently (add to your PowerShell profile)
Add-Content $PROFILE 'baq completion powershell | Out-String | Invoke-Expression'
```

### Direct Host Connection

Baqueano supports direct connection to hosts via the command line, making it easy to integrate into your existing workflow:

```bash
# Connect directly to any configured host
baq production-server
baq db-staging
baq web-01

# All direct connections are tracked in your history
# Use the TUI to see your most recently connected hosts
```

**Features of Direct Connection:**
- **Instant connection** - No TUI navigation required
- **History tracking** - All connections are recorded with timestamps
- **Error handling** - Clear messages if host doesn't exist or configuration issues
- **Config file support** - Works with custom config files using `-c` flag

### Remote Command Execution

Execute commands on remote hosts without opening an interactive shell:

```bash
# Execute a single command
baq prod-server uptime

# Execute command with arguments
baq prod-server ls -la /var/log

# Check disk usage
baq prod-server df -h

# View logs (pipe to local commands)
baq prod-server 'cat /var/log/nginx/access.log' | grep 404

# Force TTY allocation for interactive commands (sudo, vim, etc.)
baq -t prod-server sudo systemctl restart nginx
```

**Features:**
- **Exit code propagation** - Remote command exit codes are passed through
- **TTY support** - Use `-t` flag for commands requiring terminal interaction
- **Pipe-friendly** - Output can be piped to local commands for processing
- **History tracking** - Command executions are recorded in connection history

### Backup Configuration

Baqueano automatically creates backups of your SSH configuration files before making any changes to ensure your configurations are safe.

**Backup Location:**
- **Unix/Linux/macOS**: `~/.config/baqueano/backups/` (or `$XDG_CONFIG_HOME/baqueano/backups/` if set)
- **Windows**: `%APPDATA%\baqueano\backups\` (fallback: `%USERPROFILE%\.config\baqueano\backups\`)

**Key Features:**
- Automatic backup before any modification
- One backup per file (overwrites previous backup)
- Stored separately to avoid SSH Include conflicts
- Easy manual recovery if needed

**Additional Storage:**
- **Connection History**: Stored in the same config directory for persistent tracking
- **Port Forwarding History**: Saved configurations for quick reuse of common forwarding setups

**Quick Recovery:**
```bash
# Unix/Linux/macOS
cp ~/.config/baqueano/backups/config.backup ~/.ssh/config

# Windows
copy "%APPDATA%\baqueano\backups\config.backup" "%USERPROFILE%\.ssh\config"
```

### Configuration File Options

By default, Baqueano uses the standard SSH configuration file at `~/.ssh/config`. You can specify a different configuration file using the `-c` flag:

```bash
# Use custom config file in TUI mode
baq -c /path/to/custom/ssh_config

# Use custom config file with commands
baq add hostname -c /path/to/custom/ssh_config
baq edit hostname -c /path/to/custom/ssh_config
baq move hostname -c /path/to/custom/ssh_config
```

### Google Drive Sync (experimental)

Keep the same SSH host list on several machines by storing it in your Google
Drive. Hosts live in a hidden **appDataFolder** that only baq can see, so they
never mix with your regular Drive files. Only the SSH config document is
synced: **private keys are never uploaded**.

**One-time setup**

Just authorize once:

```bash
baq drive login
```

baq ships with a built-in OAuth client (compiled into the binary), so this
opens the browser to finish authorization — no ids to enter. This applies to
official releases, and to local builds made with the bundled build script or
`make` (see "Building from source" below).

<details>
<summary>Using your own OAuth client (optional)</summary>

To use your own client instead of the embedded one, provide it either as
environment variables:

```bash
export BAQUEANO_GOOGLE_CLIENT_ID=xxxxxxxx.apps.googleusercontent.com
export BAQUEANO_GOOGLE_CLIENT_SECRET=xxxxxxxx
```

or in `<config-dir>/drive_client.json`:

```json
{ "client_id": "...", "client_secret": "..." }
```

(`<config-dir>` is `%APPDATA%\baqueano` on Windows, `~/.config/baqueano` elsewhere.)
Non-interactive alternative: `baq drive login --client-id ID --client-secret SECRET`.
Precedence: env vars > `drive_client.json` > embedded client.

</details>

> **Maintainers:** the embedded client is injected at build time from
> `BAQUEANO_GOOGLE_CLIENT_ID` / `BAQUEANO_GOOGLE_CLIENT_SECRET` — GitHub Actions secrets
> for releases, or a gitignored `.env` (from `.env.example`) for local builds.
> Nothing secret is committed. Desktop-app client secrets are not treated as
> confidential by Google — they ship inside every installed app.

**Everyday use**

```bash
baq drive sync          # merge local + remote (local wins), upload the result
baq drive sync --dry-run # show what would change, write nothing
baq drive pull          # overwrite local with the remote copy (no merge)
baq drive push          # overwrite the remote with the local copy (no merge)
```

`sync` writes a backup of the current local config to the baq backup
directory before overwriting it. Remote-only hosts are appended; if a host
exists on both sides the local definition is kept.

### Advanced Features

#### Host Movement Between Config Files

Baqueano provides a powerful `move` command to relocate SSH hosts between different configuration files. **This feature requires SSH Include directives to be present in your SSH configuration.**

```bash
# Move a host to another config file (requires Include directives)
baq move my-server

# Move with custom config file (requires Include directives)
baq move my-server -c /path/to/custom/ssh_config
```

**Important Requirements:**
- **SSH Include directives must be present** in your SSH config file (either `~/.ssh/config` or the file specified with `-c`)
- The config file must contain `Include` statements referencing other SSH configuration files
- Without Include directives, the move command will display an error message

**Features:**
- **Interactive file selector** - Choose destination config file from Include directives
- **Include support** - Works seamlessly with SSH Include directives structure
- **Atomic operations** - Safe host movement with automatic backups
- **Validation** - Prevents conflicts and ensures configuration integrity
- **Error handling** - Clear messages when Include files are needed but not found

**Use Cases:**
- Reorganize hosts from main config to specialized include files
- Move development hosts to separate environment-specific configs
- Consolidate configurations for better organization

**Example Setup Required:**
Your main SSH config file must contain Include directives like:
```ssh
# ~/.ssh/config
Include ~/.ssh/config.d/*
Include work-servers.conf
Include projects/*.conf

Host personal-server
    HostName personal.example.com
    User myuser
```

#### Connectivity Status

Baqueano features asynchronous SSH connectivity checking that provides visual indicators of host availability:

**Status Indicators:**
- **Online** - SSH connection successful (shows response time)
- **Connecting** - Currently testing connectivity
- **Offline** - SSH connection failed or host unreachable
- **Unknown** - Status not yet determined

**Features:**
- **Non-blocking checks** - Status updates happen in the background
- **Response time tracking** - See connection latency for online hosts
- **Checked at startup** - Every visible host is checked when the TUI starts; press `p` to check again. Set `ping_on_start` to `false` to check only on demand
- **Error details** - Detailed error information for failed connections

#### Automatic Update Checking

Baqueano includes built-in version checking that notifies you of available updates:

**Features:**
- **Background checking** - Version check happens asynchronously, never blocking startup
- **Release notifications** - Clear indicators when updates are available
- **Pre-release detection** - Identifies beta and development versions
- **GitHub integration** - Direct links to release pages
- **Non-intrusive** - Updates don't interrupt your workflow
- **Configurable** - Can be disabled for air-gapped or offline environments

**Update notifications appear:**
- In the main TUI interface as a subtle notification
- Only when a newer stable version is available

**Disabling update checks:**

Via the CLI flag (one-time):
```bash
baq --no-update-check
```

Via `~/.config/baqueano/config.json` (persistent):
```json
{
  "check_for_updates": false
}
```

#### Port Forwarding History

Baqueano remembers your port forwarding configurations for easy reuse:

**Features:**
- **Automatic saving** - Successful forwarding setups are saved automatically
- **Quick reuse** - Previously used configurations appear as suggestions
- **Per-host history** - Forwarding history is tracked per SSH host
- **All forward types** - Supports Local (-L), Remote (-R), and Dynamic (-D) forwarding history
- **Persistent storage** - History survives application restarts

### Platform-Specific Notes

**Windows:**
- Baqueano works with the built-in OpenSSH client (Windows 10/11)
- Configuration file location: `%USERPROFILE%\.ssh\config`
- Compatible with WSL SSH configurations
- Supports the same SSH options as Unix systems

**Unix/Linux/macOS:**
- Standard SSH configuration file: `~/.ssh/config`
- Full compatibility with OpenSSH features
- Preserves file permissions automatically

## Configuration

Baqueano works directly with your standard SSH configuration file (`~/.ssh/config`). It adds special comment tags for enhanced functionality while maintaining full compatibility with standard SSH tools.

### SSH Include Support

Baqueano fully supports SSH Include directives, allowing you to organize your SSH configurations across multiple files. This is particularly useful for managing large numbers of hosts or organizing configurations by environment, project, or team.

**Include Examples:**
```ssh
# Main ~/.ssh/config file
Host personal-server
    HostName personal.example.com
    User myuser

# Include work-related configurations
Include work-servers.conf

# Include all configurations from a directory
Include projects/*

# Include with relative paths
Include ~/.ssh/configs/production.conf
```

**Organization Examples:**

*work-servers.conf:*
```ssh
# Tags: work, production
Host prod-web-01
    HostName 10.0.1.10
    User deploy
    ProxyJump bastion.company.com

# Tags: work, staging  
Host staging-api
    HostName staging-api.company.com
    User developer
```

*projects/client-alpha.conf:*
```ssh
# Tags: client, development
Host client-alpha-dev
    HostName dev.client-alpha.com
    User admin
    Port 2222
```

**Example configuration:**
Include ~/.ssh/conf.d/*

```ssh
# Tags: production, web, frontend
Host web-prod-01
    HostName 192.168.1.10
    User deploy
    Port 22
    IdentityFile ~/.ssh/production_key
    Compression yes
    ServerAliveInterval 60

# Tags: development, database
Host db-dev
    HostName dev-db.company.com
    User admin
    Port 2222
    IdentityFile ~/.ssh/dev_key
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null

# Tags: production, backend
Host backend-prod
    HostName 10.0.1.50
    User app
    Port 22
    ProxyJump bastion.company.com
    ProxyCommand ssh -W %h:%p Jumphost
    IdentityFile ~/.ssh/production_key
    Compression yes
    ServerAliveInterval 300
    BatchMode yes
```

### Supported SSH Options

Baqueano supports all standard SSH configuration options:

**Built-in Fields:**
- `HostName` - Server hostname or IP address
- `User` - Username for SSH connection
- `Port` - SSH port number
- `IdentityFile` - Path to private key file
- `ProxyJump` - Jump server for connection tunneling (e.g., `user@jumphost:port`)
- `ProxyCommand` - Jump command for connection tunneling (e.g, `ssh -W %h:%p Jumphost`)
- `Tags` - Custom tags (Baqueano extension); the special tag `hidden` hides the host from the TUI and `baq search` while keeping it connectable via `baq <host>`

**Additional SSH Options:**
You can add any valid SSH option using the "SSH Options" field in the interactive forms. Enter them in command-line format (e.g., `-o Compression=yes -o ServerAliveInterval=60`) and Baqueano will automatically convert them to the proper SSH config format.

**Common SSH Options:**
- `Compression` - Enable/disable compression (`yes`/`no`)
- `ServerAliveInterval` - Interval in seconds for keepalive messages
- `ServerAliveCountMax` - Maximum number of keepalive messages
- `StrictHostKeyChecking` - Host key verification (`yes`/`no`/`ask`)
- `UserKnownHostsFile` - Path to known hosts file
- `BatchMode` - Disable interactive prompts (`yes`/`no`)
- `ConnectTimeout` - Connection timeout in seconds
- `ControlMaster` - Connection multiplexing (`yes`/`no`/`auto`)
- `ControlPath` - Path for control socket
- `ControlPersist` - Keep connection alive duration
- `ForwardAgent` - Forward SSH agent (`yes`/`no`)
- `LocalForward` - Local port forwarding (e.g., `8080:localhost:80`)
- `RemoteForward` - Remote port forwarding
- `DynamicForward` - SOCKS proxy port forwarding

**Example usage in forms:**
```
SSH Options: -o Compression=yes -o ServerAliveInterval=60 -o StrictHostKeyChecking=no
```

This will be automatically converted to:
```ssh
    Compression yes
    ServerAliveInterval 60
    StrictHostKeyChecking no
```

### Application Configuration

Baqueano supports a configuration file to customize its behavior, including key bindings and update checking.

**Configuration File Location:**
- **Linux/macOS**: `~/.config/baqueano/config.json`
- **Windows**: `%APPDATA%\baqueano\config.json`

**Example Configuration:**
```json
{
  "check_for_updates": false,
  "focus_search_on_start": false,
  "ping_on_start": false,
  "key_bindings": {
    "quit_keys": ["q", "ctrl+c"],
    "disable_esc_quit": true
  }
}
```

**Available Options:**
- **check_for_updates**: Boolean to enable or disable the automatic update check at startup. Default: `true`. Set to `false` on air-gapped or offline machines to avoid connection delays.
- **focus_search_on_start**: Boolean to choose where the focus is when the TUI starts. Default: `true` (the search bar). Set to `false` to start on the host table; the `-s` / `--search` flag still starts on the search bar.
- **ping_on_start**: Boolean to check the status of every visible host when the TUI starts. Default: `true`. Set to `false` to check only when you press `p` (for example with many hosts, or hosts behind a slow ProxyJump).
- **quit_keys**: Array of keys that will quit the application. Default: `["q", "ctrl+c"]`
- **disable_esc_quit**: Boolean flag to disable ESC key from quitting the application. Default: `false`

**For Vim Users:**
If you frequently press ESC accidentally causing the application to quit, set `disable_esc_quit` to `true`. This will disable ESC as a quit key while preserving all other functionality.

**For Air-gapped Machines:**
If Baqueano is slow to start due to DNS timeouts when reaching GitHub, set `check_for_updates` to `false`. You can also use the `--no-update-check` CLI flag for a one-time override without editing the config file.

**Default Configuration:**
If no configuration file exists, Baqueano will automatically create one with default settings that maintain backward compatibility.

## Development

### Prerequisites

- Go 1.27+ 
- Git

### Build from Source

```bash
# Clone the repository
git clone https://github.com/Nicozilabe/baqueano.git
cd baq

# Build the binary
go build -o baq .

# Run
./baq
```

To make the experimental `baq drive` commands work without being prompted for
credentials, build with the bundled script, which embeds the OAuth client from a
gitignored `.env`:

```bash
cp .env.example .env   # then fill in your Google OAuth client id/secret
./scripts/build.sh     # macOS/Linux       -> dist/baq
scripts\build.cmd      # Windows (cmd/PS)  -> dist\baq.exe
# or, if you have make:
make build
```

On Windows, `scripts\build.cmd` is recommended because the default PowerShell
execution policy blocks plain `.ps1` files; it runs the same script with
`-ExecutionPolicy Bypass`. You can also run `scripts/build.ps1` directly via
`powershell -ExecutionPolicy Bypass -File scripts/build.ps1`.

A plain `go build` works too; `baq drive login` will simply ask for the
credentials the first time.

### Project Structure

```
baq/
├── main.go             # Application entry point
├── cmd/                # CLI commands (Cobra)
│   ├── root.go         # Root command and interactive mode
│   ├── add.go          # Add host command
│   ├── edit.go         # Edit host command
│   ├── move.go         # Move host command
│   └── search.go       # Search command
├── internal/
│   ├── config/         # SSH configuration management
│   │   └── ssh.go      # Config parsing and manipulation
│   ├── connectivity/   # SSH connectivity checking
│   │   └── ping.go     # Asynchronous SSH ping functionality
│   ├── history/        # Connection history tracking
│   │   ├── history.go  # History management and last login tracking
│   │   └── port_forward_test.go # Port forwarding history tests
│   ├── version/        # Version checking and updates
│   │   ├── version.go  # GitHub release checking and version comparison
│   │   └── version_test.go # Version parsing and comparison tests
│   ├── ui/             # Terminal UI components (Bubble Tea)
│   │   ├── tui.go      # Main TUI interface and program setup
│   │   ├── model.go    # Core TUI model and state
│   │   ├── update.go   # Message handling and state updates
│   │   ├── view.go     # UI rendering and layout
│   │   ├── table.go    # Host list table component with status indicators
│   │   ├── add_form.go # Add host form interface
│   │   ├── edit_form.go# Edit host form interface
│   │   ├── move_form.go# Move host form interface
│   │   ├── port_forward_form.go # Port forwarding setup with history
│   │   ├── styles.go   # Lip Gloss styling definitions
│   │   ├── sort.go     # Sorting and filtering logic
│   │   └── utils.go    # UI utility functions
│   └── validation/     # Input validation
│       └── ssh.go      # SSH config validation
├── images/             # Documentation assets
│   ├── logo.png        # Project logo
│   └── baq.gif        # Demo animation
├── install/            # Installation scripts
│   ├── unix.sh              # Unix/Linux/macOS installer
│   ├── windows.ps1          # Windows installer
│   ├── unix-uninstall.sh    # Unix/Linux/macOS uninstaller
│   ├── windows-uninstall.ps1# Windows uninstaller
│   └── README.md            # Installation guide
├── .github/            # GitHub configuration
│   ├── copilot-instructions.md # Development guidelines
│   └── workflows/      # CI/CD pipelines
│       └── build.yml   # Multi-platform builds
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
├── LICENSE             # MIT license
└── README.md           # Project documentation
```

### Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Go Crypto SSH](https://golang.org/x/crypto/ssh) - SSH connectivity checking

## Releases

Automated releases are built for multiple platforms:

| Platform | Architecture | Download |
|----------|-------------|----------|
| Linux | AMD64 | [baqueano_Linux_x86_64.tar.gz](https://github.com/Nicozilabe/baqueano/releases/latest/download/baqueano_Linux_x86_64.tar.gz) |
| Linux | ARM64 | [baqueano_Linux_arm64.tar.gz](https://github.com/Nicozilabe/baqueano/releases/latest/download/baqueano_Linux_arm64.tar.gz) |
| macOS | Intel | [baqueano_Darwin_x86_64.tar.gz](https://github.com/Nicozilabe/baqueano/releases/latest/download/baqueano_Darwin_x86_64.tar.gz) |
| macOS | Apple Silicon | [baqueano_Darwin_arm64.tar.gz](https://github.com/Nicozilabe/baqueano/releases/latest/download/baqueano_Darwin_arm64.tar.gz) |
| Windows | AMD64 | [baqueano_Windows_x86_64.zip](https://github.com/Nicozilabe/baqueano/releases/latest/download/baqueano_Windows_x86_64.zip) |
| Windows | ARM64 | [baqueano_Windows_arm64.zip](https://github.com/Nicozilabe/baqueano/releases/latest/download/baqueano_Windows_arm64.zip) |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Development Workflow

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Guillaume Archambault (@Gu1llaum-3)](https://github.com/Gu1llaum-3) for the original project this fork is based on
- [Charm](https://charm.sh/) for the amazing TUI libraries
- [Cobra](https://cobra.dev/) for the excellent CLI framework
- [@yimeng](https://github.com/yimeng) for contributing SSH Include directive support
- [@ldreux](https://github.com/ldreux) for contributing multi-word search functionality
- [@qingfengzxr](https://github.com/qingfengzxr) for contributing custom key bindings support
- The Go community for building such fantastic tools

---

<div align="center">

**Fork maintained by [Nicozilabe](https://github.com/Nicozilabe) — originally created by [Guillaume (@Gu1llaum-3)](https://github.com/Gu1llaum-3)**

**Star this repo if you found it useful!**

</div>
