#!/bin/sh
# Custom Clouddley CLI installer script

set -e

# Set the GitHub repository
GITHUB_REPO="clouddley/cli"

# Fetch the latest release version using GitHub API
fetch_latest_version() {
    curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | \
    grep '"tag_name":' | \
    sed -E 's/.*"([^"]+)".*/\1/'  # Keep the 'v' for the version
}

# Use the fetched version or default to a specified version if fetching fails
version=${1:-$(fetch_latest_version)}

if [ -z "$version" ]; then
    echo "Error: Unable to fetch the latest version and no version specified."
    exit 1
fi

echo "Installing Clouddley CLI version $version"

# Define installation directories
clouddley_install="${CLOUDDLEY_INSTALL:-$HOME/.clouddley}"
bin_dir="$clouddley_install/bin"
tmp_dir="$clouddley_install/tmp"
exe="$bin_dir/clouddley"

mkdir -p "$bin_dir"
mkdir -p "$tmp_dir"

# Construct the download URL
os=$(uname -s)
arch=$(uname -m)
case "$os" in
    Linux) os="Linux" ;;
    Darwin) os="macOS" ;;
    *) echo "Unsupported OS: $os"; exit 1 ;;
esac
case "$arch" in
    x86_64) arch="x86_64" ;;
    arm64) arch="arm64" ;;
    *) echo "Unsupported architecture: $arch"; exit 1 ;;
esac

# Remove the 'v' from the filename part of the URL
filename_version=$(echo "$version" | sed 's/^v//')
download_url="https://github.com/$GITHUB_REPO/releases/download/$version/clouddley_${filename_version}_${os}_${arch}.tar.gz"

# Download and extract the tarball
echo "Downloading Clouddley CLI from $download_url..."
curl -fSL --progress-bar "$download_url" -o "$tmp_dir/clouddley.tar.gz"
tar -C "$tmp_dir" -xzf "$tmp_dir/clouddley.tar.gz"
chmod +x "$tmp_dir/clouddley"

# Atomically move the executable to bin directory
mv "$tmp_dir/clouddley" "$exe"
rm -rf "$tmp_dir"  # Clean up

# Create a symlink or add the path to shell profile
if [ ! -L "$bin_dir/clouddley" ]; then
    ln -sf "$exe" "$bin_dir/clouddley"
fi

# Check installation and provide feedback
echo "Clouddley CLI was installed successfully to $exe"
if command -v clouddley >/dev/null; then
    echo "Run 'clouddley --help' to get started"
else
    case $SHELL in
    /bin/zsh) shell_profile=".zshrc" ;;
    *) shell_profile=".bash_profile" ;;
    esac
    echo "Manually add the directory to your \$HOME/$shell_profile (or similar):"
    echo "  export CLOUDDLEY_INSTALL=\"$clouddley_install\""
    echo "  export PATH=\"\$CLOUDDLEY_INSTALL/bin:\$PATH\""
    echo "Run '$exe --help' to get started"
fi
