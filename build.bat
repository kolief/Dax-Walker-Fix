@echo off
echo Building Dax Walker Fix for Windows...
echo.

REM Check if Go is installed
go version >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo ERROR: Go is not installed or not in PATH
    echo Please download and install Go from: https://golang.org/dl/
    echo.
    pause
    exit /b 1
)

echo Go found! Building executable...
echo.

REM Clean up dependencies
echo Cleaning up Go modules...
go mod tidy
if %ERRORLEVEL% neq 0 (
    echo ERROR: Failed to tidy modules
    pause
    exit /b 1
)

REM Build the executable with optimizations
echo Building optimized executable...
go build -ldflags "-s -w" -o daxwalkerfix.exe ./cmd/daxwalkerfix
if %ERRORLEVEL% neq 0 (
    echo ERROR: Build failed
    pause
    exit /b 1
)

echo.
echo Build successful! Created: daxwalkerfix.exe
echo.

REM Check if proxy file exists
if not exist "proxy.txt" (
    echo WARNING: proxy.txt not found - create this file with your SOCKS5 proxies
)

echo.
echo Required files for deployment:
echo   - daxwalkerfix.exe (built)
if exist "proxy.txt" (echo   - proxy.txt (found)) else (echo   - proxy.txt (missing - required for operation))
echo.

echo Usage:
echo   Basic usage:        daxwalkerfix.exe
echo   With debug logging: daxwalkerfix.exe -debug
echo   Custom timeout:     daxwalkerfix.exe -timeout 10
echo   Combined flags:     daxwalkerfix.exe -debug -timeout 15
echo.

echo File size:
for %%A in (daxwalkerfix.exe) do echo   daxwalkerfix.exe: %%~zA bytes
echo.

pause