@echo off
setlocal enabledelayedexpansion
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0

echo - Building Go binary...

if exist dist\pm-engine del /f /q dist\pm-engine
if not exist dist mkdir dist

go build -o dist\pm-engine ./cmd/pm-bin

if errorlevel 1 (
    echo Build failed!
    exit /b 1
)

echo - Build complete: dist\pm-engine
endlocal
