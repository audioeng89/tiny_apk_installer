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

if [ "$ALL" = true ]; then
    echo "Building Tiny APK Installer for all platforms..."

    mkdir -p build

    for p in "${platforms[@]}"; do
        IFS=',' read -r goos goarch name <<< "$p"
        
        echo "  Building $goos/$goarch..."
        CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build -ldflags="-s -w -X main.Version=$version" -o "build/tiny-apk-installer-v$version-$name" .
        
        if [ -f "build/tiny-apk-installer-v$version-$name" ]; then
            size=$(du -h "build/tiny-apk-installer-$version-$name" | cut -f1)
            echo "    -> build/tiny-apk-installer-$version-$name ($size)"
        fi
    done

    echo ""
    echo "Build complete!"
else
    echo "Building Tiny APK Installer for current platform..."

    mkdir -p build

    case "$(uname)" in
        Linux*)  platform="linux-x64" ;;
        Darwin*)  
            case "$(uname -m)" in
                x86_64) platform="darwin-x64" ;;
                arm64) platform="darwin-arm64" ;;
            esac
            ;;
        MINGW*|CYGWIN*|MSYS*) 
            platform="win-x64"
            ;;
    esac

    CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$version" -o "build/tiny-apk-installer-v$version-$platform" .

    size=$(du -h "build/tiny-apk-installer-v$version-$platform" | cut -f1)
    echo ""
    echo "Build complete!"
    echo "  Executable: build/tiny-apk-installer-v$version-$platform"
    echo "  Size: $size"
fi
