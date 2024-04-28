# Custom Clouddley CLI installer script for PowerShell

# Exit on error
$ErrorActionPreference = "Stop"

# Set the GitHub repository
$GITHUB_REPO = "clouddley/cli"

# Function to fetch the latest release version using GitHub API
function Fetch-LatestVersion {
    $apiUrl = "https://api.github.com/repos/$GITHUB_REPO/releases/latest"
    $response = Invoke-RestMethod -Uri $apiUrl
    return $response.tag_name
}

# Use the fetched version or default to a specified version if fetching fails
$version = $args[0]
if (-not $version) {
    $version = Fetch-LatestVersion
}

if (-not $version) {
    Write-Error "Error: Unable to fetch the latest version and no version specified."
    exit 1
}

Write-Host "Installing Clouddley CLI version $version"

# Define installation directories
$clouddley_install = "${env:CLOUDDLEY_INSTALL -join $env:USERPROFILE\.clouddley}"
$bin_dir = Join-Path $clouddley_install "bin"
$tmp_dir = Join-Path $clouddley_install "tmp"
$exe = Join-Path $bin_dir "clouddley.exe"

New-Item -ItemType Directory -Force -Path $bin_dir, $tmp_dir | Out-Null

# Construct the download URL
$os = if ($env:OS -eq "Windows_NT") { "Windows" } else { Write-Error "Unsupported OS: $env:OS"; exit 1 }
$arch = if ([Environment]::Is64BitOperatingSystem) { "x86_64" } else { "arm64" } # Simplified assumption

# Remove the 'v' from the filename part of the URL
$filename_version = $version -replace '^v', ''
$download_url = "https://github.com/$GITHUB_REPO/releases/download/$version/clouddley_${filename_version}_${os}_${arch}.zip"

# Download and extract the zip file
Write-Host "Downloading Clouddley CLI from $download_url..."
Invoke-WebRequest -Uri $download_url -OutFile "$tmp_dir\clouddley.zip"
Expand-Archive -Path "$tmp_dir\clouddley.zip" -DestinationPath $tmp_dir -Force

# Move the executable to the bin directory
Move-Item -Path "$tmp_dir\clouddley.exe" -Destination $exe -Force
Remove-Item -Path $tmp_dir -Recurse -Force

# Add to PATH if not already present
$envPath = [Environment]::GetEnvironmentVariable("PATH", [EnvironmentVariableTarget]::User)
if (-not $envPath.Contains($bin_dir)) {
    [Environment]::SetEnvironmentVariable("PATH", "$envPath;$bin_dir", [EnvironmentVariableTarget]::User)
    Write-Host "Added $bin_dir to your PATH"
}

# Check installation and provide feedback
Write-Host "Clouddley CLI was installed successfully to $exe"
if (Test-Path $exe) {
    Write-Host "Run 'clouddley --help' to get started"
} else {
    Write-Host "Please ensure $exe is in your PATH to use the Clouddley CLI"
}
