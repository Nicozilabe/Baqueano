# Builds baq, injecting the Google OAuth client from .env (if present) so that
# `baq drive login` works without prompting. `.env` is gitignored.
#
# Usage:  pwsh -File scripts/build.ps1
$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot

$envFile = Join-Path $root ".env"
if (Test-Path $envFile) {
    Get-Content $envFile | ForEach-Object {
        $line = $_.Trim()
        if ($line -eq "" -or $line.StartsWith("#")) { return }
        $i = $line.IndexOf("=")
        if ($i -lt 1) { return }
        $key = $line.Substring(0, $i).Trim()
        $val = $line.Substring($i + 1).Trim()
        [Environment]::SetEnvironmentVariable($key, $val, "Process")
    }
}

$version = "dev"
$ldflags = "-s -w -X github.com/Nicozilabe/baqueano/cmd.AppVersion=$version"
if ($env:BAQUEANO_GOOGLE_CLIENT_ID) {
    $ldflags += " -X github.com/Nicozilabe/baqueano/internal/drive.builtinClientID=$($env:BAQUEANO_GOOGLE_CLIENT_ID)"
}
if ($env:BAQUEANO_GOOGLE_CLIENT_SECRET) {
    $ldflags += " -X github.com/Nicozilabe/baqueano/internal/drive.builtinClientSecret=$($env:BAQUEANO_GOOGLE_CLIENT_SECRET)"
}

$outDir = Join-Path $root "dist"
New-Item -ItemType Directory -Force -Path $outDir | Out-Null
$out = Join-Path $outDir "baq.exe"

Push-Location $root
try {
    & go build -ldflags $ldflags -o $out .
    if ($LASTEXITCODE -ne 0) { throw "go build failed" }
} finally {
    Pop-Location
}

if ($env:BAQUEANO_GOOGLE_CLIENT_ID) {
    Write-Host "Built $out (with embedded Google client)"
} else {
    Write-Host "Built $out (no .env / client id; 'baq drive login' will ask for credentials)"
}
