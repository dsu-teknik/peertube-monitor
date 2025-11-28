# Windows Installer

This directory contains the WiX Toolset configuration for building a Windows MSI installer.

## Quick Start

From the project root:

```powershell
./build-installer.ps1
```

This will:
1. Build the Windows executable (`peertube-monitor.exe`)
2. Generate an MSI installer with a graphical configuration wizard
3. Output to `build/PeerTubeMonitor-1.0.0.msi`

## Installer Features

The MSI installer provides:

- **Guided Configuration Wizard** with three dialogs:
  1. **PeerTube Settings** – server URL, username, password
  2. **Folder Settings** – watch, done, and failed folder paths
  3. **Video Defaults** – category, licence, language, privacy, description, and advanced options

- **Windows Service Installation**:
  - Automatically installs as "PeerTubeMonitor" service
  - Configured to start automatically on boot
  - Runs under LocalSystem account
  - Credentials stored securely in service environment variables

- **Configuration File Generation**:
  - Creates `config.json` in Program Files with all settings
  - Credentials stored in environment variables (not in config file)
  - Service automatically finds config at install location

- **Clean Uninstallation**:
  - Stops and removes Windows service
  - Removes program files
  - Watch/done/failed folders are NOT deleted (data preservation)

## File Structure

- **`Product.wxs`** – WiX source file defining the installer
  - Package metadata and upgrade logic
  - Custom UI dialogs for configuration
  - Service installation and environment variables
  - Custom action to generate config.json from user input

- **`installer.wixproj`** – MSBuild project file
  - References WiX SDK v5
  - Includes UI and Util extensions

- **`build-installer.ps1`** – Build script
  - Builds Go executable
  - Compiles WiX project
  - Outputs versioned MSI

## Requirements

**Development machine:**
- Windows 10/11 or Windows Server 2019+
- .NET SDK 6.0 or later
- WiX Toolset v5: `dotnet tool install --global wix`
- Go 1.21+ (for building the executable)

**Target machine (installer only):**
- Windows 10/11 or Windows Server 2019+
- .NET Runtime (usually already installed)

## Building

### Standard Build

```powershell
# From project root
./build-installer.ps1
```

### Custom Version

```powershell
./build-installer.ps1 -Version "2.1.5"
```

### Skip Executable Build

If you've already built `peertube-monitor.exe`:

```powershell
./build-installer.ps1 -SkipBuild
```

## Testing

Install with verbose logging to troubleshoot issues:

```powershell
msiexec /i "build\PeerTubeMonitor-1.0.0.msi" /l*v install.log
```

Check the log file for detailed installation steps and any errors.

## Configuration Parameters

Default values set in `Product.wxs`:

| Parameter | Default | Description |
|-----------|---------|-------------|
| PEERTUBE_URL | https://peertube.sandum.net | PeerTube server URL |
| PEERTUBE_USERNAME | (empty) | Username for authentication |
| PEERTUBE_PASSWORD | (empty) | Password for authentication |
| WATCH_PATH | C:\Videos\Upload | Folder to monitor for new videos |
| DONE_PATH | C:\Videos\Done | Where successful uploads go |
| FAILED_PATH | C:\Videos\Failed | Where failed uploads go |
| VIDEO_CATEGORY | 5 | Default video category |
| VIDEO_LICENCE | 9 | Default video licence |
| VIDEO_LANGUAGE | da | Default language code |
| VIDEO_PRIVACY | 1 | Privacy (1=Public, 2=Unlisted, 3=Private) |
| VIDEO_DESCRIPTION | Automatically uploaded | Default description |
| SETTLE_TIME | 5 | Seconds to wait before processing |
| MAX_RETRIES | 3 | Upload retry attempts |
| COMMENTS_ENABLED | 1 | Enable comments (1=yes, 0=no) |
| DOWNLOAD_ENABLED | 0 | Enable downloads (1=yes, 0=no) |

## Customization

### Change Default Values

Edit `Product.wxs` and modify the `<Property>` elements:

```xml
<Property Id="PEERTUBE_URL" Value="https://your-server.com" />
<Property Id="VIDEO_LANGUAGE" Value="en" />
```

### Add New Configuration Options

1. Add a property in `Product.wxs`
2. Add UI control in the appropriate dialog
3. Update the `GenerateConfigFile` custom action to include the new property

### Modify Upgrade GUID

The `UpgradeCode` GUID identifies the product family. Change it only if creating a completely different product:

```xml
<Package UpgradeCode="YOUR-NEW-GUID-HERE">
```

## Troubleshooting

**Build fails with "wix not found":**
```powershell
dotnet tool install --global wix
```

**Build fails with ".NET SDK not found":**
- Install .NET SDK from https://dot.net

**Service doesn't start after installation:**
- Check Windows Event Viewer → Application logs
- Verify credentials are correct
- Check that PeerTube server URL is accessible

**Config file not created:**
- Check install log for custom action errors
- Verify INSTALLFOLDER permissions

## Architecture Notes

The installer uses a VBScript custom action (`GenerateConfigFile`) to dynamically create the config.json file from MSI properties. This approach ensures:

- User input is captured during installation
- No template file dependencies
- JSON structure matches application expectations
- Boolean values converted correctly (true/false not 1/0)

Service environment variables are set via `ServiceConfig/Environment` elements, which integrate with Windows Service Manager.
