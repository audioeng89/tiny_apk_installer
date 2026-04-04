# Build script for Tiny APK Installer

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

version=$(grep -oP 'Version\s*=\s*"\K[^"]+' version.go)
if [ -z "$version" ]; then
    version="dev"
fi

echo "Building Tiny APK Installer v$version..."

ALL=false
if [ "$1" = "-a" ] || [ "$1" = "--all" ]; then
    ALL=true
fi

platforms=(
    "windows,amd64,win-x64"
    "windows,arm64,win-arm64"
    "darwin,amd64,darwin-x64"
    "darwin,arm64,darwin-arm64"
    "linux,amd64,linux-x64"
    "linux,arm64,linux-arm64"
)

mkdir -p build

if [ "$ALL" = true ]; then
    echo "Building Tiny APK Installer for all platforms..."

    for p in "${platforms[@]}"; do
        IFS=',' read -r goos goarch name <<< "$p"
        
        echo "  Building $goos/$goarch..."
        CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build -ldflags="-s -w -X main.Version=$version" -o "build/tiny-apk-installer-v$version-$name" .
        
        if [ -f "build/tiny-apk-installer-v$version-$name" ]; then
            size=$(du -h "build/tiny-apk-installer-v$version-$name" | cut -f1)
            echo "    -> build/tiny-apk-installer-v$version-$name ($size)"
        fi
    done
else
    echo "Building Tiny APK Installer for current platform..."
    CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$version" -o "build/tiny-apk-installer" .
    size=$(du -h "build/tiny-apk-installer" | cut -f1)
    echo "  Executable: build/tiny-apk-installer"
    echo "  Size: $size"
fi

echo ""
echo "Build complete!"
