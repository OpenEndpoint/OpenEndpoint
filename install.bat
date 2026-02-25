@echo off
REM OpenEndpoint One-Click Installer for Windows (Batch)
REM Usage: curl -fsSL https://raw.githubusercontent.com/OpenEndpoint/OpenEndpoint/main/install.bat | cmd

echo ========================================
echo   OpenEndpoint Windows Installer
echo ========================================
echo.

set INSTALL_DIR=%LOCALAPPDATA%\OpenEndpoint
set CONFIG_DIR=%USERPROFILE%\.openendpoint
set REPO=OpenEndpoint/OpenEndpoint

REM Detect architecture
if "%PROCESSOR_ARCHITECTURE%"=="AMD64" (
    set ARCH=amd64
) else (
    echo Unsupported architecture: %PROCESSOR_ARCHITECTURE%
    exit /b 1
)

echo Detected architecture: %ARCH%

REM Get latest version
echo Checking latest version...
for /f "tokens=*" %%a in ('powershell -Command "(Invoke-RestMethod -Uri 'https://api.github.com/repos/%REPO%/releases/latest').tag_name"') do set VERSION=%%a

if not defined VERSION (
    echo Failed to get latest version
    exit /b 1
)

echo Latest version: %VERSION%
echo.

REM Create directories
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"
if not exist "%CONFIG_DIR%" mkdir "%CONFIG_DIR%"
if not exist "%CONFIG_DIR%\data" mkdir "%CONFIG_DIR%\data"

REM Download
echo Downloading OpenEndpoint %VERSION% for Windows %ARCH%...
set DOWNLOAD_URL=https://github.com/%REPO%/releases/download/%VERSION%/openep-%VERSION%-windows-%ARCH%.zip
set ZIP_FILE=%TEMP%\openep-%RANDOM%.zip

powershell -Command "Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile '%ZIP_FILE%' -UseBasicParsing"

if errorlevel 1 (
    echo Download failed
    exit /b 1
)

REM Extract
echo Extracting...
powershell -Command "Expand-Archive -Path '%ZIP_FILE%' -DestinationPath '%TEMP%\openep-extract' -Force"

REM Move binary
if exist "%TEMP%\openep-extract\openep-windows-amd64.exe" (
    move /Y "%TEMP%\openep-extract\openep-windows-amd64.exe" "%INSTALL_DIR%\openep.exe"
) else (
    move /Y "%TEMP%\openep-extract\openep.exe" "%INSTALL_DIR%\openep.exe"
)

REM Cleanup
del /F /Q "%ZIP_FILE%"
rmdir /S /Q "%TEMP%\openep-extract"

echo.
echo OpenEndpoint installed to %INSTALL_DIR%\openep.exe

REM Add to PATH
echo Adding to PATH...
setx PATH "%PATH%;%INSTALL_DIR%" >nul 2>&1

REM Create config
echo Creating sample configuration...
(
echo server:
echo   host: "0.0.0.0"
echo   port: 8080
echo.
echo storage:
echo   type: "flatfile"
echo   path: ".\data"
echo.
echo logging:
echo   level: "info"
echo   format: "json"
echo.
echo buckets:
echo   - name: "my-bucket"
echo     region: "us-east-1"
echo   - name: "uploads"
echo     region: "us-east-1"
echo.
echo credentials:
echo   - access_key: "demo-access-key"
echo     secret_key: "demo-secret-key"
echo     buckets:
echo       - "my-bucket"
echo       - "uploads"
) > "%CONFIG_DIR%\config.yaml"

REM Create sample data
if not exist "%CONFIG_DIR%\sample-data" mkdir "%CONFIG_DIR%\sample-data"
echo Hello from OpenEndpoint! > "%CONFIG_DIR%\sample-data\hello.txt"
(
echo {
echo   "message": "Welcome to OpenEndpoint",
echo   "version": "1.0.0",
echo   "features": [
echo     "S3-compatible API",
echo     "Multi-platform support",
echo     "Easy deployment"
echo   ]
echo }
) > "%CONFIG_DIR%\sample-data\sample.json"

REM Create start script
(
echo @echo off
echo echo Starting OpenEndpoint...
echo "%INSTALL_DIR%\openep.exe" --config "%CONFIG_DIR%\config.yaml"
echo pause
) > "%CONFIG_DIR%\start.bat"

echo.
echo ========================================
echo   Installation Complete!
echo ========================================
echo.
echo To start OpenEndpoint:
echo   1. Run: %CONFIG_DIR%\start.bat
echo   2. Or: %INSTALL_DIR%\openep.exe --config %CONFIG_DIR%\config.yaml
echo.
echo Sample data location:
echo   %CONFIG_DIR%\sample-data\
echo.
echo API Endpoint:
echo   http://localhost:8080
echo.
echo Default credentials:
echo   Access Key: demo-access-key
echo   Secret Key: demo-secret-key
echo.
echo Note: Restart your terminal for 'openep' command to work
echo.

pause
