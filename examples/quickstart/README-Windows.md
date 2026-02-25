# OpenEndpoint QuickStart for Windows

One-command setup to get OpenEndpoint running on Windows.

## Prerequisites

- Windows 10/11 (64-bit)
- PowerShell 5.1+ or PowerShell Core 7+
- Or: Docker Desktop for Windows

## Method 1: PowerShell Installer (Recommended)

### Quick Install
```powershell
# Run in PowerShell as Administrator
Invoke-Expression (Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/OpenEndpoint/OpenEndpoint/main/install.ps1').Content
```

### Or download and run manually:
```powershell
# Download
Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/OpenEndpoint/OpenEndpoint/main/install.ps1' -OutFile 'install.ps1'

# Run
.\install.ps1
```

## Method 2: Batch Installer

```cmd
# Download using curl (Windows 10 1803+)
curl -fsSL -o install.bat https://raw.githubusercontent.com/OpenEndpoint/OpenEndpoint/main/install.bat

# Run
install.bat
```

## Method 3: Docker (Easiest)

```powershell
# Clone repository
git clone https://github.com/OpenEndpoint/OpenEndpoint.git
cd OpenEndpoint\examples\quickstart

# Start with Docker Compose
docker-compose -f docker-compose.windows.yml up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f openendpoint
```

## After Installation

### Start OpenEndpoint
```powershell
# Using the start script
$env:USERPROFILE\.openendpoint\start.bat

# Or directly
$env:LOCALAPPDATA\OpenEndpoint\openep.exe --config $env:USERPROFILE\.openendpoint\config.yaml
```

### Access the API
- **S3 API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

### Default Credentials
- **Access Key**: `demo-access-key`
- **Secret Key**: `demo-secret-key`

## Test with AWS CLI for Windows

```powershell
# Install AWS CLI if not installed
# https://aws.amazon.com/cli/

# Configure
aws configure --profile openendpoint
# AWS Access Key ID: demo-access-key
# AWS Secret Access Key: demo-secret-key
# Default region: us-east-1
# Default output: json

# List buckets
aws --profile openendpoint --endpoint-url http://localhost:8080 s3 ls

# Create a bucket
aws --profile openendpoint --endpoint-url http://localhost:8080 s3 mb s3://my-test-bucket

# Upload a file
aws --profile openendpoint --endpoint-url http://localhost:8080 s3 cp hello.txt s3://my-test-bucket/

# List objects
aws --profile openendpoint --endpoint-url http://localhost:8080 s3 ls s3://my-test-bucket/
```

## Test with PowerShell

```powershell
# Health check
Invoke-RestMethod -Uri 'http://localhost:8080/health'

# List buckets (requires AWS signature)
# Use AWS CLI or a tool like Postman
```

## Windows Service (Optional)

To run OpenEndpoint as a Windows service:

```powershell
# Download NSSM (Non-Sucking Service Manager)
# https://nssm.cc/download

# Install as service
nssm install OpenEndpoint "$env:LOCALAPPDATA\OpenEndpoint\openep.exe"
nssm set OpenEndpoint AppParameters "--config $env:USERPROFILE\.openendpoint\config.yaml"
nssm start OpenEndpoint
```

## Troubleshooting

### Port already in use
```powershell
# Find process using port 8080
netstat -ano | findstr :8080

# Kill process
taskkill /PID <PID> /F
```

### Firewall blocking
```powershell
# Add firewall rule (run as Administrator)
netsh advfirewall firewall add rule name="OpenEndpoint" dir=in action=allow protocol=tcp localport=8080
```

### Permission denied
```powershell
# Run PowerShell as Administrator
# Or use -ExecutionPolicy Bypass flag:
PowerShell -ExecutionPolicy Bypass -File install.ps1
```

## Stop and Clean Up

### Binary installation:
```powershell
# Stop OpenEndpoint (Ctrl+C in the running window)

# Remove data
Remove-Item -Path "$env:USERPROFILE\.openendpoint\data" -Recurse -Force

# Uninstall
Remove-Item -Path "$env:LOCALAPPDATA\OpenEndpoint" -Recurse -Force
Remove-Item -Path "$env:USERPROFILE\.openendpoint" -Recurse -Force
```

### Docker:
```powershell
docker-compose -f docker-compose.windows.yml down
docker-compose -f docker-compose.windows.yml down -v  # Remove data too
```
