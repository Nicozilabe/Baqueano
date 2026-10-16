# Baqueano Windows Installation Script
# Usage: 
#   Online:  irm https://raw.githubusercontent.com/Nicozilabe/baqueano/main/install/windows.ps1 | iex
#   Local:   .\install\windows.ps1 -LocalBinary ".\baq.exe"

param(
    [string]$InstallDir = "$env:LOCALAPPDATA\baqueano",
    [switch]$Force = $false,
    [string]$LocalBinary = ""
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

Write-Info "🚀 Installing Baqueano - SSH Manager"
Write-Info ""

# Check if Baqueano is already installed
$existingBaqueano = Get-Command baq -ErrorAction SilentlyContinue
if ($existingBaqueano -and -not $Force) {
    $currentVersion = & baq --version 2>$null | Select-String "version" | ForEach-Object { $_.ToString().Split()[-1] }
    Write-Warning "Baqueano is already installed (version: $currentVersion)"
    $response = Read-Host "Do you want to continue with the installation? (y/N)"
    if ($response -ne "y" -and $response -ne "Y") {
        Write-Info "Installation cancelled."
        exit 0
    }
}

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
Write-Info "Detected platform: Windows ($arch)"

# Check if using local binary
if ($LocalBinary -ne "") {
    if (-not (Test-Path $LocalBinary)) {
        Write-Error "Local binary not found: $LocalBinary"
        exit 1
    }
    
    Write-Info "Using local binary: $LocalBinary"
    $targetPath = "$InstallDir\baq.exe"
    
    # Create installation directory
    if (-not (Test-Path $InstallDir)) {
        Write-Info "Creating installation directory: $InstallDir"
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }
    
    # Copy local binary
    Write-Info "Installing binary to: $targetPath"
    Copy-Item -Path $LocalBinary -Destination $targetPath -Force
    
} else {
    # Online installation
    Write-Info "Starting online installation..."
    
    # Get latest version
    Write-Info "Fetching latest version..."
    try {
        $latestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/Nicozilabe/baqueano/releases/latest"
        $latestVersion = $latestRelease.tag_name
        Write-Info "Target version: $latestVersion"
    } catch {
        Write-Error "Failed to fetch version information"
        exit 1
    }

    # Download binary
    # Map architecture to match GoReleaser format
    $goreleaserArch = if ($arch -eq "amd64") { "x86_64" } else { "i386" }
    
    # GoReleaser format: baqueano_Windows_x86_64.zip
    $fileName = "baqueano_Windows_$goreleaserArch.zip"
    $downloadUrl = "https://github.com/Nicozilabe/baqueano/releases/download/$latestVersion/$fileName"
    $tempFile = "$env:TEMP\$fileName"

    Write-Info "Downloading $fileName..."
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile
    } catch {
        Write-Error "Download failed"
        exit 1
    }

    # Verify the downloaded archive against the published checksums
    $checksumUrl = "https://github.com/Nicozilabe/baqueano/releases/download/$latestVersion/checksums.txt"
    $checksumFile = "$env:TEMP\baq-checksums.txt"

    Write-Info "Verifying checksum..."
    try {
        Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumFile
    } catch {
        Write-Error "Failed to download checksums"
        exit 1
    }

    $expected = (Get-Content $checksumFile |
        Where-Object { $_ -match "\s$([regex]::Escape($fileName))$" } |
        ForEach-Object { ($_ -split '\s+')[0] } |
        Select-Object -First 1)

    if (-not $expected) {
        Write-Error "No checksum found for $fileName"
        exit 1
    }

    $actual = (Get-FileHash -Path $tempFile -Algorithm SHA256).Hash
    if ($actual -ne $expected.ToUpper()) {
        Write-Error "Checksum verification failed for $fileName"
        Write-Error "Expected: $expected"
        Write-Error "Actual:   $actual"
        exit 1
    }
    Write-Info "Checksum verified for $fileName"

    # Create installation directory
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    # Extract archive
    Write-Info "Extracting..."
    try {
        Expand-Archive -Path $tempFile -DestinationPath $env:TEMP -Force
        # GoReleaser extracts the binary as just "baq.exe", not with platform suffix
        $extractedBinary = "$env:TEMP\baq.exe"
        $targetPath = "$InstallDir\baq.exe"
        
        Move-Item -Path $extractedBinary -Destination $targetPath -Force
    } catch {
        Write-Error "Extraction failed"
        exit 1
    }

    # Clean up
    Remove-Item $tempFile -Force
    Remove-Item $checksumFile -Force -ErrorAction SilentlyContinue
}

# Check PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$InstallDir*") {
    Write-Warning "The directory $InstallDir is not in your PATH."
    Write-Info "Adding to user PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
    Write-Info "Please restart your terminal to use the 'baq' command."
}

Write-Info ""
Write-Info "✅ Baqueano successfully installed to: $targetPath"
Write-Info "You can now use the 'baq' command!"

# Verify installation
if (Test-Path $targetPath) {
    Write-Info ""
    Write-Info "Verifying installation..."
    & $targetPath --version
}
