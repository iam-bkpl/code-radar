#!/bin/bash
set -e

REPO="iam-bkpl/code-radar"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="code-radar"

# Get latest release version
get_latest_version() {
    curl -s "https://api.github.com/repos/${REPO}/releases/latest" | \
        grep '"tag_name"' | \
        sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'
}

# Get architecture
get_arch() {
    local arch=$(uname -m)
    case $arch in
        x86_64)  echo "x86_64" ;;
        aarch64) echo "arm64" ;;
        arm64)   echo "arm64" ;;
        *)       echo "amd64" ;;
    esac
}

# Get OS
get_os() {
    local os=$(uname -s)
    case $os in
        Linux)  echo "Linux" ;;
        Darwin) echo "Darwin" ;;
        *)      echo "$os" ;;
    esac
}

# Download and install
install() {
    local version=$1
    local os=$(get_os)
    local arch=$(get_arch)
    
    echo "📦 Installing code-radar ${version}"
    echo "   OS: ${os}, Arch: ${arch}"
    
    # Download
    local filename="code-radar_${os}_${arch}.tar.gz"
    local url="https://github.com/${REPO}/releases/download/${version}/${filename}"
    
    echo "   Downloading ${url}..."
    curl -L -o "/tmp/${filename}" "${url}"
    
    # Extract
    echo "   Extracting..."
    tar xzf "/tmp/${filename}" -C /tmp
    
    # Install
    echo "   Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
    sudo mv "/tmp/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    
    # Cleanup
    rm -f "/tmp/${filename}"
    
    # macOS Gatekeeper fix
    if [ "$os" = "Darwin" ]; then
        echo "   Removing macOS quarantine attribute..."
        xattr -cr "${INSTALL_DIR}/${BINARY_NAME}" 2>/dev/null || true
    fi
    
    echo ""
    echo "✅ code-radar ${version} installed successfully!"
    echo ""
    echo "   Run 'code-radar --help' to get started"
    echo ""
}

# Main
main() {
    echo "🚀 Code Radar Installer"
    echo ""
    
    # Check if already installed
    if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        local current_version=$(${INSTALL_DIR}/${BINARY_NAME} --version 2>/dev/null | awk '{print $2}' || echo "unknown")
        echo "   Current version: ${current_version}"
    fi
    
    # Get latest version
    echo "   Checking for latest release..."
    local version=$(get_latest_version)
    
    if [ -z "$version" ]; then
        echo "❌ Could not determine latest version"
        echo "   Check https://github.com/${REPO}/releases manually"
        exit 1
    fi
    
    echo "   Latest version: ${version}"
    echo ""
    
    # Install
    install "$version"
}

main "$@"
