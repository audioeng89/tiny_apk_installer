# Build script for Tiny APK Installer

param(
    [switch]$All
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

$versionRegex = 'Version\s*=\s*"([^"]+)"'
$versionFile = Get-Content "version.go" -Raw
if ($versionFile -match $versionRegex) {
    $version = $matches[1]
} else {
    $version = "dev"
}

Write-Host "Building Tiny APK Installer v$version..."

$platforms = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Name = "win-x64" },
    @{ GOOS = "windows"; GOARCH = "arm64"; Name = "win-arm64" },
    @{ GOOS = "darwin"; GOARCH = "amd64"; Name = "darwin-x64" },
    @{ GOOS = "darwin"; GOARCH = "arm64"; Name = "darwin-arm64" },
    @{ GOOS = "linux"; GOARCH = "amd64"; Name = "linux-x64" },
    @{ GOOS = "linux"; GOARCH = "arm64"; Name = "linux-arm64" }
)

if (-not (Test-Path "build")) {
    New-Item -ItemType Directory -Path "build" | Out-Null
}

if ($All) {
    Write-Host "Building Tiny APK Installer for all platforms..."

    foreach ($p in $platforms) {
        $env:GOOS = $p.GOOS
        $env:GOARCH = $p.GOARCH
        $env:CGO_ENABLED = "0"

        $ext = if ($p.GOOS -eq "windows") { ".exe" } else { "" }
        $output = "build/tiny-apk-installer-v$($version)-$($p.Name)$ext"

        Write-Host "  Building $($p.GOOS)/$($p.GOARCH)..."
        go build -ldflags="-s -w -X main.Version=$version" -o $output .

        if (Test-Path $output) {
            $size = (Get-Item $output).Length / 1MB
            $sizeStr = "{0:N2} MB" -f $size
            Write-Host "    -> $output ($sizeStr)"
        }
    }
} else {
    Write-Host "Building Tiny APK Installer for current platform..."

    $env:CGO_ENABLED = "0"
    $output = "build/tiny-apk-installer.exe"

    go build -ldflags="-s -w -X main.Version=$version" -o $output .

    $exe = Get-Item $output
    Write-Host "  Executable: $output"
    Write-Host ("  Size: {0:N2} MB" -f ($exe.Length / 1MB))
}

Write-Host ""
Write-Host "Build complete!"
