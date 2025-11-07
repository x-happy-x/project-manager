#!/usr/bin/env pwsh
$ErrorActionPreference = "Stop"

Write-Host "🚀 Installing Project Manager (pm)..." -ForegroundColor Cyan

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$OS = "windows"
$BinaryName = "pm.exe"

# GitHub repository
$Repo = "yourusername/project-manager"

try {
    Write-Host "📦 Detecting latest release..." -ForegroundColor Yellow

    # Get latest release info
    $LatestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    $Asset = $LatestRelease.assets | Where-Object { $_.name -like "*${OS}_${Arch}*" } | Select-Object -First 1

    if ($null -eq $Asset) {
        throw "No pre-built binary found for $OS-$Arch"
    }

    # Download binary
    Write-Host "⬇️  Downloading from $($Asset.browser_download_url)..." -ForegroundColor Yellow
    $TempDir = New-Item -ItemType Directory -Path ([System.IO.Path]::GetTempPath()) -Name ([System.IO.Path]::GetRandomFileName())
    $BinaryPath = Join-Path $TempDir.FullName $BinaryName

    Invoke-WebRequest -Uri $Asset.browser_download_url -OutFile $BinaryPath

} catch {
    Write-Host "⚠️  Could not download pre-built binary: $_" -ForegroundColor Yellow
    Write-Host "⚠️  Building from source..." -ForegroundColor Yellow

    # Check if Go is installed
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Host "❌ Go is not installed. Please install Go 1.21+ from https://golang.org/dl/" -ForegroundColor Red
        exit 1
    }

    # Clone and build
    $TempDir = New-Item -ItemType Directory -Path ([System.IO.Path]::GetTempPath()) -Name ([System.IO.Path]::GetRandomFileName())
    Set-Location $TempDir.FullName

    git clone "https://github.com/$Repo.git"
    Set-Location "project-manager"

    go build -o $BinaryName ./cmd/pm
    $BinaryPath = Join-Path (Get-Location) $BinaryName
}

# Install binary
$InstallDir = Join-Path $env:USERPROFILE "bin"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

$FinalPath = Join-Path $InstallDir $BinaryName
Copy-Item $BinaryPath $FinalPath -Force

Write-Host "✅ Installed to $FinalPath" -ForegroundColor Green

# Add to PATH if not already there
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "📝 Adding to PATH..." -ForegroundColor Yellow
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    $env:Path = "$env:Path;$InstallDir"
    Write-Host "✅ Added to PATH (restart terminal or run: `$env:Path = [System.Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path','User'))" -ForegroundColor Green
}

# Initialize pm
Write-Host "⚙️  Initializing pm..." -ForegroundColor Yellow
& $FinalPath init

Write-Host ""
Write-Host "🎉 Installation complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Quick start:" -ForegroundColor Cyan
Write-Host "  pm add <path>          # Add a project" -ForegroundColor White
Write-Host "  pm list                # List all projects" -ForegroundColor White
Write-Host "  pm open <name>         # Open a project" -ForegroundColor White
Write-Host "  pm --help              # Show all commands" -ForegroundColor White

# Cleanup
Remove-Item -Recurse -Force $TempDir.FullName -ErrorAction SilentlyContinue
