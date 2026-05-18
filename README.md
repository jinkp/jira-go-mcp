# jira-go-mcp

Go-based MCP server for Jira release management, plus setup helpers for OpenCode, Claude Code, and Claude Desktop.

## What it does

- Exposes Jira release workflows over **MCP stdio**.
- Lets you register `jira-mcp` into supported AI clients with:
  - `jira-mcp setup opencode`
  - `jira-mcp setup claude`
  - `jira-mcp setup claude-desktop`
- Includes an interactive **Bubble Tea** setup wizard:
  - `jira-mcp tui`
- Includes a read-only install checker:
  - `jira-mcp doctor`

## Supported clients

- OpenCode
- Claude Code
- Claude Desktop

## Install

### Option 1: download a release binary

Download the correct archive for your platform from the [GitHub Releases page](https://github.com/jinkp/jira-go-mcp/releases).

Expected artifact names:

- `jira-go-mcp_Windows_x86_64.zip`
- `jira-go-mcp_Linux_x86_64.tar.gz`
- `jira-go-mcp_Darwin_arm64.tar.gz`

After extracting, put `jira-mcp` (or `jira-mcp.exe`) somewhere in your `PATH`.

### Option 2: install with `irm` (PowerShell / Windows)

```powershell
$version = "v0.1.1"
$asset = "jira-go-mcp_0.1.1_windows_amd64.zip"
$url = "https://github.com/jinkp/jira-go-mcp/releases/download/$version/$asset"
$tmp = Join-Path $env:TEMP $asset
$dest = Join-Path $env:USERPROFILE "bin\jira-go-mcp"
New-Item -ItemType Directory -Force -Path $dest | Out-Null
irm $url -OutFile $tmp
Expand-Archive -LiteralPath $tmp -DestinationPath $dest -Force
```

Then add the destination folder to your `PATH`, for example:

```powershell
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$dest*") {
  [Environment]::SetEnvironmentVariable("Path", "$userPath;$dest", "User")
}
```

### Option 3: install with `curl` (Linux / macOS)

#### Linux amd64

```bash
VERSION="v0.1.1"
ASSET="jira-go-mcp_0.1.1_linux_amd64.tar.gz"
curl -Lo /tmp/$ASSET https://github.com/jinkp/jira-go-mcp/releases/download/$VERSION/$ASSET
mkdir -p "$HOME/.local/bin/jira-go-mcp"
tar -xzf /tmp/$ASSET -C "$HOME/.local/bin/jira-go-mcp"
```

#### macOS arm64

```bash
VERSION="v0.1.1"
ASSET="jira-go-mcp_0.1.1_darwin_arm64.tar.gz"
curl -Lo /tmp/$ASSET https://github.com/jinkp/jira-go-mcp/releases/download/$VERSION/$ASSET
mkdir -p "$HOME/.local/bin/jira-go-mcp"
tar -xzf /tmp/$ASSET -C "$HOME/.local/bin/jira-go-mcp"
```

Then add it to your shell `PATH`:

```bash
export PATH="$HOME/.local/bin/jira-go-mcp:$PATH"
```

### Option 4: build from source

```bash
git clone https://github.com/jinkp/jira-go-mcp.git
cd jira-go-mcp
go build -o jira-mcp .
```

## Quick start

### 1. Configure Jira credentials

Copy the example file and set your values:

```bash
cp .env.example .env
```

Required environment variables:

```env
JIRA_BASE_URL=https://your-company.atlassian.net
JIRA_EMAIL=you@company.com
JIRA_API_TOKEN=replace_me
```

Optional:

```env
JIRA_API_VERSION=3
DEFAULT_DONE_STATUSES=Done,Closed,Released
DEFAULT_CRITICAL_PRIORITIES=Highest,Critical
```

### 2. Register the MCP server in your client

#### OpenCode

```bash
jira-mcp setup opencode --global
```

#### Claude Code

```bash
jira-mcp setup claude --global
```

#### Claude Desktop

```bash
jira-mcp setup claude-desktop
```

### 3. Or use the interactive wizard

```bash
jira-mcp tui
```

### 4. Check installation status

```bash
jira-mcp doctor
```

### 5. Run the MCP server

```bash
jira-mcp mcp
```

## CLI overview

```bash
jira-mcp version
jira-mcp mcp
jira-mcp setup opencode --global
jira-mcp setup opencode --local
jira-mcp setup claude --global
jira-mcp setup claude --local
jira-mcp setup claude-desktop
jira-mcp tui
jira-mcp doctor
```

## MCP capabilities

Current release-management tools include:

- create release
- update release
- mark release as released
- archive release
- list releases by project
- get issues for a release
- generate release notes
- validate a release for deploy

## Safety notes

- Do **not** commit real Jira credentials.
- `.env*` files are gitignored by default.
- Local AI metadata and coverage artifacts are also ignored.
- The MCP stdio command must stay silent on stdout before server startup.

## Development

Run tests:

```bash
go test ./...
```

Check version:

```bash
jira-mcp version
```

## Release automation

This repository includes:

- **GoReleaser** configuration in `.goreleaser.yaml`
- **GitHub Actions** workflow in `.github/workflows/release.yml`

Pushing a version tag like `v0.1.1` publishes cross-platform release archives automatically.

## Roadmap

- richer README examples per MCP client
- live Jira integration smoke tests
- polished installation docs with screenshots/GIFs for the TUI wizard

## License

MIT
