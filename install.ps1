# OpenEndpoint One-Click Installer for Windows (PowerShell)
# Usage: Invoke-Expression (Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/OpenEndpoint/OpenEndpoint/main/install.ps1').Content

param(
    [string]$InstallDir = "$env:LOCALAPPDATA\OpenEndpoint",
    [string]$ConfigDir = "$env:USERPROFILE\.openendpoint"
)

$ErrorActionPreference = "Stop"

# Colors
$Green = "`e[32m"
$Red = "`e[31m"
$Yellow = "`e[33m"
$Reset = "`e[0m"

$Repo = "OpenEndpoint/OpenEndpoint"

function Write-Color($Color, $Message) {
    Write-Host "$Color$Message$Reset"
}

function Get-LatestVersion {
    $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    return $response.tag_name
}

function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default {
            Write-Color $Red "Unsupported architecture: $arch"
            exit 1
        }
    }
}

function Download-AndInstall {
    param($Version, $Arch)

    $platform = "windows-$Arch"
    $binary = "openep-${platform}.exe"
    $downloadUrl = "https://github.com/$Repo/releases/download/$Version/openep-${Version}-windows-${Arch}.zip"

    Write-Color $Yellow "Downloading OpenEndpoint $Version for $platform..."

    # Create temp directory
    $tmpDir = Join-Path $env:TEMP "openendpoint-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    try {
        # Download
        $zipPath = Join-Path $tmpDir "openep.zip"
        Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing

        # Extract
        Expand-Archive -Path $zipPath -DestinationPath $tmpDir -Force

        # Create install directory
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

        # Move binary
        $sourceBinary = Join-Path $tmpDir "openep-windows-amd64.exe"
        if (-not (Test-Path $sourceBinary)) {
            # Try alternative name
            $sourceBinary = Join-Path $tmpDir "openep.exe"
        }
        $destBinary = Join-Path $InstallDir "openep.exe"
        Move-Item -Path $sourceBinary -Destination $destBinary -Force

        Write-Color $Green "OpenEndpoint installed to $destBinary"

        # Add to PATH if not already there
        $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($currentPath -notlike "*$InstallDir*") {
            Write-Color $Yellow "Adding $InstallDir to your PATH..."
            [Environment]::SetEnvironmentVariable("Path", "$currentPath;$InstallDir", "User")
            Write-Color $Green "PATH updated. Please restart your terminal."
        }

        return $destBinary
    }
    finally {
        # Cleanup
        Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Setup-SampleConfig {
    Write-Color $Yellow "Setting up sample configuration..."

    New-Item -ItemType Directory -Path $ConfigDir -Force | Out-Null

    $configContent = @"
server:
  host: "0.0.0.0"
  port: 8080

storage:
  type: "flatfile"
  path: ".\data"

logging:
  level: "info"
  format: "json"

buckets:
  - name: "my-bucket"
    region: "us-east-1"
  - name: "uploads"
    region: "us-east-1"

credentials:
  - access_key: "demo-access-key"
    secret_key: "demo-secret-key"
    buckets:
      - "my-bucket"
      - "uploads"
"@

    $configPath = Join-Path $ConfigDir "config.yaml"
    $configContent | Out-File -FilePath $configPath -Encoding UTF8

    Write-Color $Green "Sample configuration created at $configPath"
}

function Create-SampleData {
    Write-Color $Yellow "Creating sample data..."

    $sampleDir = Join-Path $ConfigDir "sample-data"
    New-Item -ItemType Directory -Path $sampleDir -Force | Out-Null

    # Sample text file
    "Hello from OpenEndpoint!" | Out-File -FilePath (Join-Path $sampleDir "hello.txt") -Encoding UTF8

    # Sample JSON
    $jsonContent = @"
{
  "message": "Welcome to OpenEndpoint",
  "version": "1.0.0",
  "features": [
    "S3-compatible API",
    "Multi-platform support",
    "Easy deployment"
  ]
}
"@
    $jsonContent | Out-File -FilePath (Join-Path $sampleDir "sample.json") -Encoding UTF8

    Write-Color $Green "Sample data created in $sampleDir"
}

function Create-StartScript {
    $startScript = @"
@echo off
echo Starting OpenEndpoint...
"$InstallDir\openep.exe" --config "$ConfigDir\config.yaml"
pause
"@

    $startPath = Join-Path $ConfigDir "start.bat"
    $startScript | Out-File -FilePath $startPath -Encoding ASCII

    Write-Color $Green "Start script created at $startPath"
}

# Main
Write-Color $Green "========================================"
Write-Color $Green "  OpenEndpoint Windows Installer"
Write-Color $Green "========================================"
Write-Host ""

$arch = Get-Architecture
Write-Host "Detected architecture: " -NoNewline
Write-Color $Yellow $arch

$version = Get-LatestVersion
if (-not $version) {
    Write-Color $Red "Failed to get latest version"
    exit 1
}
Write-Host "Latest version: " -NoNewline
Write-Color $Yellow $version

Download-AndInstall -Version $version -Arch $arch
Setup-SampleConfig
Create-SampleData
Create-StartScript

Write-Host ""
Write-Color $Green "========================================"
Write-Color $Green "  Installation Complete!"
Write-Color $Green "========================================"
Write-Host ""
Write-Host "To start OpenEndpoint:"
Write-Host "  1. Run: " -NoNewline
Write-Color $Yellow "$ConfigDir\start.bat"
Write-Host "  2. Or: " -NoNewline
Write-Color $Yellow "openep.exe --config $ConfigDir\config.yaml"
Write-Host ""
Write-Host "Sample data location:"
Write-Host "  " -NoNewline
Write-Color $Yellow "$ConfigDir\sample-data\"
Write-Host ""
Write-Host "API Endpoint:"
Write-Host "  " -NoNewline
Write-Color $Yellow "http://localhost:8080"
Write-Host ""
Write-Host "Default credentials:"
Write-Host "  Access Key: " -NoNewline
Write-Color $Yellow "demo-access-key"
Write-Host "  Secret Key: " -NoNewline
Write-Color $Yellow "demo-secret-key"
Write-Host ""
Write-Host "Note: If 'openep' command not found, restart your terminal or run:"
Write-Color $Yellow "  $InstallDir\openep.exe --help"
Write-Host ""
