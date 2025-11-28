# PeerTube Monitor

[![Latest Release](https://img.shields.io/github/v/release/dsu-teknik/peertube-monitor)](https://github.com/dsu-teknik/peertube-monitor/releases/latest)
[![Downloads](https://img.shields.io/github/downloads/dsu-teknik/peertube-monitor/total)](https://github.com/dsu-teknik/peertube-monitor/releases)

Automatic video uploader for PeerTube. Monitors a folder for new video files and automatically uploads them to your PeerTube instance.

## Features

- **Automatic monitoring** – Watches a specified folder for new video files
- **Smart file detection** – Waits for files to finish copying/downloading before uploading
- **Automatic upload** – Uploads videos to PeerTube with configurable metadata
- **Success handling** – Moves successful uploads to a "done" folder or deletes them
- **Failure handling** – Moves failed uploads to a "failed" folder with retry logic
- **Cross-platform** – Runs on Windows, Linux, and macOS

## Requirements

- Go 1.21 or later (for building)
- PeerTube instance with valid credentials

## Installation

### Windows MSI Installer (Recommended for Windows)

**[Download Latest Windows Installer (MSI)](https://github.com/dsu-teknik/peertube-monitor/releases/latest)**

The installer provides:
- **Guided setup** – Configure all settings through the installation wizard
- **Automatic service installation** – Runs as a Windows service on startup
- **Secure credential storage** – Credentials stored in service environment (not in config file)
- **Pre-configured folders** – Watch, done, and failed folders created automatically

During installation, you'll configure:
1. PeerTube server URL, username, and password
2. Folder locations for watching, done, and failed uploads
3. Video defaults (category, licence, language, privacy, description)
4. Advanced settings (settle time, retry count, comments/downloads)

After installation, the service starts automatically and begins monitoring your watch folder.

### On Fedora/Linux

```bash
# Install Go
sudo dnf install golang

# Navigate to the project directory
cd peertube-monitor

# Download dependencies
go mod tidy

# Build for Linux
go build -o peertube-monitor ./cmd/monitor

# Or build for Windows
GOOS=windows GOARCH=amd64 go build -o peertube-monitor.exe ./cmd/monitor
```

### Building Windows Installer Locally

**Prerequisites:**
- .NET SDK (for WiX v5)
- WiX Toolset v5: `dotnet tool install --global wix`

```powershell
# Build with default version (1.0.0)
./build-installer.ps1

# Build with custom version
./build-installer.ps1 -Version "1.2.3"
```

The MSI will be created in the `installer/` directory.

### Building for Windows from Linux

```bash
# Build Windows executable
GOOS=windows GOARCH=amd64 go build -o peertube-monitor.exe ./cmd/monitor

# The resulting .exe can be copied to a Windows machine
```

### Building on Windows

```powershell
# Install Go from https://go.dev/dl/

# Build
go build -o peertube-monitor.exe ./cmd/monitor
```

## Configuration

1. Copy the example configuration:
   ```bash
   cp configs/config.example.json config.json
   ```

2. Edit `config.json` with your settings:

```json
{
  "peertube": {
    "url": "https://your-peertube-instance.com",
    "username": "",
    "password": "",
    "defaults": {
      "category": "Sports",
      "licence": "Public Domain Dedication",
      "language": "da",
      "privacy": "Public",
      "description": "Automatically uploaded",
      "tags": [],
      "downloadEnabled": false,
      "commentsEnabled": true,
      "waitTranscoding": false,
      "nsfw": false
    }
  },
  "watcher": {
    "watchPath": "C:\\Videos\\Upload",
    "donePath": "C:\\Videos\\Done",
    "failedPath": "C:\\Videos\\Failed",
    "videoExtensions": [".mp4", ".webm", ".mkv", ".avi", ".mov", ".flv"],
    "settleTime": 5,
    "maxRetries": 3
  },
  "logging": {
    "logFile": "peertube-monitor.log",
    "verbose": true
  }
}
```

### Configuration Options

#### PeerTube Settings
- **url** – Your PeerTube instance URL (can use `PEERTUBE_URL` env var)
- **username** – Your PeerTube username (can use `PEERTUBE_USERNAME` env var)
- **password** – Your PeerTube password (can use `PEERTUBE_PASSWORD` env var)
- **defaults.category** – Default video category (string name or number ID, e.g., `"Sports"` or `5`)
- **defaults.licence** – Default license (string name or number ID, e.g., `"Public Domain Dedication"` or `7`)
- **defaults.language** – Language code (e.g., "da", "en")
- **defaults.privacy** – Privacy level (string name or number ID, e.g., `"Public"` or `1`)
  - Available privacy levels: `"Public"` (1), `"Unlisted"` (2), `"Private"` (3), `"Internal"` (4), `"Password protected"` (5)
- **defaults.downloadEnabled** – Allow video downloads
- **defaults.commentsEnabled** – Enable comments

**Note:** The application fetches available categories, licences, and privacy levels from your PeerTube instance at startup. You can use either human-readable names (case-insensitive) or numeric IDs. If you provide an invalid value, the error message will list all available options.

#### Watcher Settings
- **watchPath** – Folder to monitor for new videos
- **donePath** – Where to move successful uploads (empty = delete)
- **failedPath** – Where to move failed uploads (empty = rename with .failed)
- **videoExtensions** – File extensions to monitor
- **settleTime** – Seconds to wait for file to stop changing
- **maxRetries** – Upload retry attempts before marking as failed

### Environment Variables (Recommended for Services)

For production deployments, especially when running as a Windows service, you can provide credentials via environment variables instead of storing them in the config file:

**Supported environment variables:**
- `PEERTUBE_URL` – Override PeerTube instance URL
- `PEERTUBE_USERNAME` – PeerTube username
- `PEERTUBE_PASSWORD` – PeerTube password

Environment variables take precedence over values in `config.json`. This allows you to:
- Keep sensitive credentials out of config files
- Store only non-sensitive settings (paths, metadata) in `config.json`
- Manage credentials securely via Windows service environment

**Example config.json without credentials:**
```json
{
  "peertube": {
    "url": "https://peertube.sandum.net",
    "username": "",
    "password": "",
    "defaults": { ... }
  },
  "watcher": { ... }
}
```

Then set environment variables in your service definition (see Windows Service section below).

## Usage

```bash
# Run with default config.json
./peertube-monitor

# Run with custom config
./peertube-monitor -config /path/to/config.json

# Show version
./peertube-monitor -version
```

### Running as a Service (Windows)

**Option 1: MSI Installer (Recommended)**

The MSI installer automatically sets up and starts the Windows service. No manual configuration needed.

**Option 2: Manual Installation with NSSM**

You can use NSSM (Non-Sucking Service Manager) to run this as a Windows service:

```powershell
# Download NSSM from https://nssm.cc/
nssm install PeerTubeMonitor "C:\Tools\peertube-monitor.exe"
nssm set PeerTubeMonitor AppDirectory "C:\Tools"
nssm set PeerTubeMonitor AppParameters "-config config.json"
nssm set PeerTubeMonitor DisplayName "PeerTube Upload Monitor"
nssm set PeerTubeMonitor Description "Automatically uploads videos to PeerTube"

# Set environment variables for credentials (recommended)
nssm set PeerTubeMonitor AppEnvironmentExtra PEERTUBE_USERNAME=your-username PEERTUBE_PASSWORD=your-password

nssm start PeerTubeMonitor
```

**Note:** With this setup, `config.json` in `C:\Tools\` contains only non-sensitive settings (paths, metadata defaults), while credentials are stored securely in the service's environment variables.

### Running as a Service (Linux)

Create `/etc/systemd/system/peertube-monitor.service`:

```ini
[Unit]
Description=PeerTube Monitor
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/peertube-monitor
Environment="PEERTUBE_USERNAME=your-username"
Environment="PEERTUBE_PASSWORD=your-password"
ExecStart=/path/to/peertube-monitor -config config.json
Restart=always

[Install]
WantedBy=multi-user.target
```

Then:
```bash
sudo systemctl daemon-reload
sudo systemctl enable peertube-monitor
sudo systemctl start peertube-monitor
```

## How It Works

1. **Monitoring** – The application watches the specified folder for new video files
2. **Settling** – When a new file is detected, it waits for the configured settle time to ensure the file is completely written
3. **Upload** – The video is uploaded to PeerTube with the configured metadata (video name is derived from filename)
4. **Success** – On successful upload, the file is moved to the done folder or deleted
5. **Failure** – On upload failure, the upload is retried up to maxRetries times, then moved to the failed folder

## Building Releases with GitHub Actions

The project includes a GitHub Actions workflow that automatically builds Windows MSI installers without requiring a local Windows environment or WiX license (WiX is open source).

### Creating a Release

**Option 1: Version Tag (Recommended)**

```bash
# Create and push a version tag
git tag v1.0.0
git push origin v1.0.0
```

The workflow will automatically:
- Build the Windows executable
- Create the MSI installer
- Publish a GitHub release with both files attached

**Option 2: Manual Workflow Dispatch**

1. Go to the GitHub repository
2. Click **Actions** → **Build Windows Installer**
3. Click **Run workflow**
4. Enter a version number (optional)
5. Click **Run workflow**

The MSI and executable will be available as downloadable artifacts.

### Workflow Features

- Runs on GitHub's free Windows runners
- Installs .NET SDK and WiX Toolset v5 automatically
- Builds both the executable and MSI installer
- Uploads artifacts for every build
- Creates GitHub releases for tagged versions
- No local Windows environment or WiX license required

## Project Structure

```
peertube-monitor/
├── .github/
│   └── workflows/
│       └── build-installer.yml   # GitHub Actions workflow
├── cmd/monitor/                  # Main application entry point
│   └── main.go
├── pkg/
│   ├── config/                   # Configuration handling
│   │   └── config.go
│   ├── peertube/                 # PeerTube API client
│   │   └── client.go
│   └── watcher/                  # File monitoring and handling
│       ├── watcher.go
│       └── handler.go
├── installer/                    # WiX installer source
│   └── PeerTubeMonitor.wxs
├── configs/
│   └── config.example.json
├── build-installer.ps1           # PowerShell build script
└── README.md
```

## Troubleshooting

**Authentication fails**
- Verify your PeerTube URL, username, and password
- Ensure your PeerTube instance is accessible

**Files not being detected**
- Check that watchPath exists and is readable
- Verify file extensions match videoExtensions in config
- Increase settleTime if large files are being processed too early

**Uploads fail**
- Check log file for detailed error messages
- Verify file size doesn't exceed PeerTube instance limits
- Ensure proper network connectivity

## License

MIT License
