$ErrorActionPreference = "Stop"

$repo = "yoanbernabeu/bugsnag-cli"
$binary = "bugsnag.exe"

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Write-Error "Unsupported: 32-bit systems are not supported"
    exit 1
}

# Get latest release
Write-Host "Fetching latest release..."
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
$tag = $release.tag_name
$version = $tag.TrimStart("v")

$archive = "bugsnag-cli_${version}_windows_${arch}.zip"
$url = "https://github.com/$repo/releases/download/$tag/$archive"

# Download and extract
$tmpDir = Join-Path $env:TEMP "bugsnag-cli-install"
New-Item -ItemType Directory -Force -Path $tmpDir | Out-Null

Write-Host "Downloading bugsnag-cli $tag for windows/$arch..."
Invoke-WebRequest -Uri $url -OutFile (Join-Path $tmpDir $archive)
Expand-Archive -Path (Join-Path $tmpDir $archive) -DestinationPath $tmpDir -Force

# Install to user's local bin
$installDir = Join-Path $env:LOCALAPPDATA "bugsnag-cli"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Move-Item -Force (Join-Path $tmpDir $binary) (Join-Path $installDir $binary)

# Clean up
Remove-Item -Recurse -Force $tmpDir

# Add to PATH if not already there
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
    $env:Path = "$env:Path;$installDir"
    Write-Host "Added $installDir to your PATH."
}

Write-Host ""
Write-Host "bugsnag-cli $tag installed to $installDir\$binary"
Write-Host ""
Write-Host "Get started:"
Write-Host "  bugsnag configure --api-token YOUR_TOKEN"
Write-Host "  bugsnag organizations list"
