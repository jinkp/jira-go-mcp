param(
    [string]$Version = "latest",
    [string]$InstallDir = "$HOME\bin"
)

$ErrorActionPreference = "Stop"

$owner = "jinkp"
$repo = "jira-go-mcp"
$binName = "jira-mcp.exe"

$archMap = @{
    "AMD64" = "amd64"
    "X64"   = "amd64"
    "ARM64" = "arm64"
}

$machineArch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToUpperInvariant()
if (-not $archMap.ContainsKey($machineArch)) {
    throw "Unsupported architecture: $machineArch"
}

$arch = $archMap[$machineArch]

if ($Version -eq "latest") {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/releases/latest"
    $Version = $release.tag_name
    if (-not $Version) {
        throw "Could not resolve latest release version"
    }
}

$versionNoV = $Version.TrimStart('v')
$asset = "${repo}_${versionNoV}_windows_${arch}.zip"
$url = "https://github.com/$owner/$repo/releases/download/$Version/$asset"

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$tmpZip = Join-Path $env:TEMP $asset
$tmpExtract = Join-Path $env:TEMP ("jira-go-mcp-" + [guid]::NewGuid().ToString("N"))

try {
    Invoke-RestMethod -Uri $url -OutFile $tmpZip
    New-Item -ItemType Directory -Force -Path $tmpExtract | Out-Null
    Expand-Archive -LiteralPath $tmpZip -DestinationPath $tmpExtract -Force
    Copy-Item -LiteralPath (Join-Path $tmpExtract $binName) -Destination (Join-Path $InstallDir $binName) -Force
}
finally {
    if (Test-Path -LiteralPath $tmpZip) { Remove-Item -LiteralPath $tmpZip -Force }
    if (Test-Path -LiteralPath $tmpExtract) { Remove-Item -LiteralPath $tmpExtract -Recurse -Force }
}

Write-Host "Installed $binName to $InstallDir"

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
    Write-Host "Added $InstallDir to your user PATH. Restart your shell to pick it up."
} else {
    Write-Host "$InstallDir is already present in your user PATH."
}
