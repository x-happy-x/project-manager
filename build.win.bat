@echo off
setlocal enabledelayedexpansion

echo - Building Go binary...

if exist dist\pm-engine.exe del /f /q dist\pm-engine.exe
if not exist dist mkdir dist

go build -o dist\pm-engine.exe ./cmd/pm-bin

if errorlevel 1 (
    echo Build failed!
    exit /b 1
)

echo - Build complete: dist\pm-engine.exe
endlocal
