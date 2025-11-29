#!/usr/bin/env pwsh
# Build script for PeerTube Monitor Windows installer
# Requires: Go 1.21+, WiX Toolset v5+

param(
    [string]$Version = "",
    [switch]$SkipBuild,
    [switch]$Sign
)

$ErrorActionPreference = "Stop"

# Get version from git if not specified
if ([string]::IsNullOrEmpty($Version)) {
    try {
        $Version = & git describe --tags --always --dirty 2>$null
        if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrEmpty($Version)) {
            $Version = "dev"
        }
    }
    catch {
        $Version = "dev"
    }
}

# Configuration
$ProjectRoot = $PSScriptRoot
$BuildDir = Join-Path $ProjectRoot "build"
$InstallerDir = Join-Path $ProjectRoot "installer"
$OutputMsi = Join-Path $BuildDir "PeerTubeMonitor-$Version.msi"

Write-Host "=== PeerTube Monitor Installer Build ===" -ForegroundColor Cyan
Write-Host "Version: $Version" -ForegroundColor Green

# Create build directory
if (-not (Test-Path $BuildDir)) {
    New-Item -ItemType Directory -Path $BuildDir | Out-Null
}

# Step 1: Build the Go executable
if (-not $SkipBuild) {
    Write-Host "`nStep 1: Building Windows executable..." -ForegroundColor Yellow

    $env:GOOS = "windows"
    $env:GOARCH = "amd64"

    $ExePath = Join-Path $ProjectRoot "peertube-monitor.exe"

    # Get git commit hash
    try {
        $Commit = & git rev-parse --short HEAD 2>$null
        if ($LASTEXITCODE -ne 0) {
            $Commit = "unknown"
        }
    }
    catch {
        $Commit = "unknown"
    }

    & go build -ldflags "-s -w -X main.version=$Version -X main.commit=$Commit" -o $ExePath "$ProjectRoot\cmd\monitor"

    if ($LASTEXITCODE -ne 0) {
        Write-Error "Go build failed with exit code $LASTEXITCODE"
        exit 1
    }

    Write-Host "  ✓ Built: $ExePath" -ForegroundColor Green
} else {
    Write-Host "`nStep 1: Skipping build (using existing executable)" -ForegroundColor Yellow
}

# Step 2: Verify WiX Toolset is installed
Write-Host "`nStep 2: Checking WiX Toolset..." -ForegroundColor Yellow

try {
    $wixVersion = & dotnet tool list -g | Select-String "wix"
    if ($wixVersion) {
        Write-Host "  ✓ WiX Toolset found: $wixVersion" -ForegroundColor Green
    }
    else {
        Write-Host "  ! WiX not found, installing..." -ForegroundColor Yellow
        & dotnet tool install --global wix
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Failed to install WiX Toolset"
            exit 1
        }
    }

    # Add extensions to WiX cache (required for v6)
    Write-Host "  Adding WiX extensions..." -ForegroundColor Yellow
    & wix extension add WixToolset.UI.wixext --global 2>$null
    & wix extension add WixToolset.Util.wixext --global 2>$null
}
catch {
    Write-Error "Failed to check/install WiX Toolset: $_"
    exit 1
}

# Step 3: Build MSI
Write-Host "`nStep 3: Building MSI installer..." -ForegroundColor Yellow

Push-Location $InstallerDir
try {
    # Call wix build directly instead of using MSBuild/SDK
    # Pass version as a preprocessor variable
    & wix build -arch x64 -ext WixToolset.UI.wixext -ext WixToolset.Util.wixext -d ProductVersion="$Version" -out "$OutputMsi" Product.wxs

    if ($LASTEXITCODE -ne 0) {
        Write-Error "MSI build failed with exit code $LASTEXITCODE"
        exit 1
    }

    Write-Host "  ✓ MSI created: $OutputMsi" -ForegroundColor Green
} finally {
    Pop-Location
}

# Step 4: Sign MSI (optional)
if ($Sign) {
    Write-Host "`nStep 4: Signing MSI..." -ForegroundColor Yellow

    # You'll need to configure your signing certificate
    # Example using signtool.exe:
    # & signtool sign /f "path\to\cert.pfx" /p "password" /tr http://timestamp.digicert.com /td sha256 /fd sha256 $OutputMsi

    Write-Host "  ! Signing not configured - implement in build script" -ForegroundColor Yellow
}

# Step 5: Display results
Write-Host "`n=== Build Complete ===" -ForegroundColor Cyan
Write-Host "MSI Installer: $OutputMsi" -ForegroundColor Green

$MsiSize = (Get-Item $OutputMsi).Length / 1MB
Write-Host "Size: $([math]::Round($MsiSize, 2)) MB" -ForegroundColor Green

Write-Host "`nTo test installation:" -ForegroundColor Yellow
Write-Host "  msiexec /i `"$OutputMsi`" /l*v install.log" -ForegroundColor White

Write-Host "`nTo uninstall:" -ForegroundColor Yellow
Write-Host "  msiexec /x `"$OutputMsi`"" -ForegroundColor White
