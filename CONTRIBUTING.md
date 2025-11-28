# Contributing to PeerTube Monitor

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git

### Getting Started

```bash
# Clone the repository
git clone https://github.com/yourusername/peertube-monitor.git
cd peertube-monitor

# Install dependencies
go mod tidy

# Build
go build -o peertube-monitor ./cmd/monitor

# Run tests
go test ./...
```

## Code Style

- Use spaces for indentation (not tabs)
- Follow standard Go formatting (`gofmt`)
- Single-statement blocks don't need braces
- Put `else`, `else if`, and error checks on new lines

## Project Architecture

The project follows clean architecture principles:

- **cmd/monitor/** – Application entry point, orchestration
- **pkg/config/** – Configuration loading and validation
- **pkg/peertube/** – PeerTube API client implementation
- **pkg/watcher/** – File monitoring and upload handling

## Making Changes

1. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes and test thoroughly

3. Commit with clear, concise messages:
   ```bash
   git commit -m "Add support for custom thumbnails"
   ```

4. Push and create a pull request:
   ```bash
   git push origin feature/your-feature-name
   ```

## Release Process

### Creating a New Release

1. **Update version numbers** (if applicable in code or documentation)

2. **Create and push a version tag:**
   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```

3. **GitHub Actions automatically:**
   - Builds the Windows executable
   - Creates the MSI installer using WiX Toolset
   - Publishes a GitHub release with downloadable assets

4. **Verify the release:**
   - Check the Actions tab for build status
   - Download and test the MSI installer
   - Update release notes if needed

### Manual Build (Testing)

If you want to test the installer build without creating a release:

1. Go to **Actions** → **Build Windows Installer**
2. Click **Run workflow**
3. Enter a test version number
4. Download the artifact to test

### Version Numbering

Follow semantic versioning (semver):
- **Major** (v2.0.0): Breaking changes
- **Minor** (v1.3.0): New features, backwards compatible
- **Patch** (v1.2.4): Bug fixes, backwards compatible

## Windows Installer Development

The MSI installer is built using WiX Toolset v5. Source files are in `installer/`.

**Local development (Windows with WiX installed):**

```powershell
# Build with custom version
./build-installer.ps1 -Version "1.2.3"

# Skip rebuilding the executable
./build-installer.ps1 -SkipBuild
```

**Testing without Windows:**

Use the GitHub Actions workflow to build and download the MSI as an artifact.

## Testing

Before submitting a pull request:

1. Run all tests: `go test ./...`
2. Test the application with a real PeerTube instance
3. Verify file watching, uploading, and error handling
4. Check log output for any issues

## Reporting Issues

When reporting bugs, include:
- PeerTube Monitor version
- Operating system
- PeerTube instance version
- Configuration (sanitize credentials)
- Log output with verbose mode enabled
- Steps to reproduce

## Questions?

Open an issue for questions or discussion about features and changes.
