# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PeerTube Monitor is an automatic video uploader for PeerTube. It monitors a folder for new video files and automatically uploads them to a PeerTube instance. Written in Go 1.21+, it's cross-platform and can run as a service on Windows or Linux.

## Build and Run Commands

```bash
# Download dependencies
go mod tidy

# Build for current platform
go build -o peertube-monitor ./cmd/monitor

# Build for Windows from Linux
GOOS=windows GOARCH=amd64 go build -o peertube-monitor.exe ./cmd/monitor

# Run with default config.json
./peertube-monitor

# Run with custom config
./peertube-monitor -config /path/to/config.json

# Show version
./peertube-monitor -version
```

## Windows Installer (MSI)

The project includes a WiX Toolset installer for Windows deployments.

**Prerequisites:**
- .NET SDK (for WiX v5)
- WiX Toolset v5: `dotnet tool install --global wix`

**Build MSI installer:**
```powershell
# Build with default version (1.0.0)
./build-installer.ps1

# Build with custom version
./build-installer.ps1 -Version "1.2.3"

# Skip rebuilding executable
./build-installer.ps1 -SkipBuild
```

**Installer Features:**
- Guided UI for all configuration parameters during installation
- Automatic Windows service installation and startup
- Credentials stored in service environment variables (not in config file)
- Creates config.json with all settings from installation wizard
- Folders are created by the service on first run

**Installation dialogs:**
1. PeerTube server URL, username, and password
2. Watch, done, and failed folder paths
3. Video upload defaults (category, licence, language, privacy, description)
4. Advanced settings (settle time, max retries, comments/downloads enabled)

**Manual installation/uninstallation:**
```powershell
# Install with verbose logging
msiexec /i "PeerTubeMonitor-1.0.0.msi" /l*v install.log

# Uninstall
msiexec /x "PeerTubeMonitor-1.0.0.msi"
```

## Architecture

The application follows a clean architecture with three main packages:

### pkg/config
- Loads configuration from JSON files
- Supports environment variable overrides for credentials (`PEERTUBE_URL`, `PEERTUBE_USERNAME`, `PEERTUBE_PASSWORD`)
- Validates configuration and creates required directories
- Path normalization to absolute paths

### pkg/peertube
- `Client` handles PeerTube API interactions
- OAuth authentication flow (gets client credentials, then access token)
- Video upload via multipart form data
- 30-minute HTTP timeout for large uploads

### pkg/watcher
- `Watcher` uses fsnotify to monitor a directory for video files
- Implements "settle time" pattern: waits for files to stop changing before processing
- Tracks pending files with modification time and size checks
- Scans for existing files on startup
- `UploadHandler` manages upload attempts, retry logic, and file movement

### cmd/monitor
- Application entry point in main.go
- Orchestrates configuration loading, PeerTube authentication, watcher setup, and graceful shutdown
- Logging setup (file or stdout, with optional verbose mode)

## Key Flows

**File Detection and Upload:**
1. Watcher detects new video file (CREATE or WRITE event)
2. File is added to pending state with settle timer
3. After settle time, file is checked – if still changing, timer resets
4. Once stable, UploadHandler processes the file
5. Video name is derived from filename (without extension)
6. Upload to PeerTube with configured defaults
7. On success: move to donePath or delete
8. On failure: retry up to maxRetries, then move to failedPath or rename with .failed extension

**Authentication:**
- First call to `/api/v1/oauth-clients/local` to get client credentials
- Then POST to `/api/v1/users/token` with username/password to get access token
- Token stored in client for subsequent uploads

## Configuration Structure

All configuration is in a single JSON file with three sections:
- `peertube`: PeerTube instance URL, credentials, and video upload defaults
- `watcher`: File paths, extensions to monitor, settle time, and retry settings
- `logging`: Log file path and verbosity

Environment variables take precedence over config file values for credentials (recommended for service deployments).

## Important Conventions

- Use spaces for indentation (not tabs)
- Single-statement blocks don't need braces in Go
- Relative paths in config are converted to absolute paths
- File movement uses rename, falling back to copy+delete if cross-filesystem
- Retry count is tracked per file path in memory
