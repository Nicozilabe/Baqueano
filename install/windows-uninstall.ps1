# SSHM Windows Uninstallation Script
# Usage:
#   Online:  irm https://raw.githubusercontent.com/Nicozilabe/sshm/main/install/windows-uninstall.ps1 | iex
#   Local:   .\install\windows-uninstall.ps1

param(
    [string]$InstallDir = "$env:LOCALAPPDATA\sshm",
    [switch]$Force = $false,
    [switch]$Purge = $false
)

$ErrorActionPreference = "Stop"

# Colors for output
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

function Write-Info { Write-ColorOutput Green $args }
function Write-Warning { Write-ColorOutput Yellow $args }
function Write-Error { Write-ColorOutput Red $args }

Write-Info "🗑️  Uninstalling SSHM - SSH Manager"
Write-Info ""

# Resolve the installed binary, preferring the configured install directory
$targetPath = "$InstallDir\sshm.exe"
$binaryPath = $null
if (Test-Path $targetPath) {
    $binaryPath = $targetPath
} else {
    $existingSSHM = Get-Command sshm -ErrorAction SilentlyContinue
    if ($existingSSHM) {
        $binaryPath = $existingSSHM.Source
    }
}

if (-not $binaryPath) {
    Write-Warning "SSHM does not appear to be installed (checked $InstallDir and PATH)."
}

# Confirm before doing anything destructive
if (-not $Force -and $binaryPath) {
    $response = Read-Host "Do you want to uninstall SSHM? (y/N)"
    if ($response -ne "y" -and $response -ne "Y") {
        Write-Info "Uninstallation cancelled."
        exit 0
    }
}

# Remove the binary
if ($binaryPath) {
    Write-Info "Removing binary: $binaryPath"
    Remove-Item -Path $binaryPath -Force
    if (Test-Path $binaryPath) {
        Write-Error "Failed to remove $binaryPath"
        exit 1
    }
    Write-Info "Removed $binaryPath"
}

# Remove the install directory if it is now empty
if (Test-Path $InstallDir) {
    $remaining = Get-ChildItem -Path $InstallDir -Force -ErrorAction SilentlyContinue
    if (-not $remaining) {
        Write-Info "Removing empty install directory: $InstallDir"
        Remove-Item -Path $InstallDir -Force -Recurse
    } else {
        Write-Warning "Install directory not empty, leaving it in place: $InstallDir"
    }
}

# Remove the install directory from the user PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -and ($userPath -like "*$InstallDir*")) {
    Write-Info "Removing $InstallDir from your PATH..."
    $newPath = ($userPath -split ';' | Where-Object { $_ -and ($_.TrimEnd('\') -ne $InstallDir.TrimEnd('\')) }) -join ';'
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Info "Please restart your terminal for the PATH change to take effect."
}

# Optionally remove the configuration directory
if ($Purge) {
    $configDir = if ($env:APPDATA) { "$env:APPDATA\sshm" } else { "$env:USERPROFILE\.config\sshm" }
    if (Test-Path $configDir) {
        Write-Info "Removing configuration directory: $configDir"
        Remove-Item -Path $configDir -Force -Recurse
        Write-Info "Removed $configDir"
    } else {
        Write-Warning "Configuration directory not found: $configDir"
    }
}

Write-Info ""
Write-Info "✅ SSHM was uninstalled successfully."

if (-not $Purge) {
    $configDir = if ($env:APPDATA) { "$env:APPDATA\sshm" } else { "$env:USERPROFILE\.config\sshm" }
    if (Test-Path $configDir) {
        Write-Warning "Your configuration was kept at $configDir"
        Write-Warning "Run with -Purge to remove it as well."
    }
}
