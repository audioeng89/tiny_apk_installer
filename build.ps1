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

if ($All) {
    Write-Host "Building Tiny APK Installer for all platforms..."

    if (-not (Test-Path "build")) {
        New-Item -ItemType Directory -Path "build" | Out-Null
    }

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

    Write-Host ""
    Write-Host "Build complete!"
} else {
    $os = $env:OS
    $arch = $env:PROCESSOR_ARCHITECTURE

    if ($os -eq "Windows_NT") {
        if ($arch -eq "AMD64") {
            $goos = "windows"
            $goarch = "amd64"
            $platform = "win-x64"
        } elseif ($arch -eq "ARM64") {
            $goos = "windows"
            $goarch = "arm64"
            $platform = "win-arm64"
        }
    } elseif ($os -eq "Unix") {
        $goos = "linux"
        $goarch = if ($arch -eq "x86_64") { "amd64" } elseif ($arch -eq "aarch64") { "arm64" } else { "amd64" }
        $platform = "linux-x64"
    }

    Write-Host "Building Tiny APK Installer for current platform ($platform)..."

    if (-not (Test-Path "build")) {
        New-Item -ItemType Directory -Path "build" | Out-Null
    }

    $env:CGO_ENABLED = "0"
    $env:GOOS = $goos
    $env:GOARCH = $goarch

    $ext = if ($goos -eq "windows") { ".exe" } else { "" }
    $output = "build/tiny-apk-installer-$version-$platform$ext"

    go build -ldflags="-s -w -X main.Version=$version" -o $output .

    $exe = Get-Item $output
    Write-Host ""
    Write-Host "Build complete!"
    Write-Host "  Executable: $output"
    Write-Host ("  Size: {0:N2} MB" -f ($exe.Length / 1MB))
}
