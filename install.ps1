#Requires -Version 5.1
$ErrorActionPreference = 'Stop'

$Repo = 'gumlet/cli'
$Binary = 'gumlet.exe'
$Version = $env:GUMLET_VERSION
$InstallDir = $env:GUMLET_INSTALL_DIR

if (-not $InstallDir) {
    $InstallDir = Join-Path $env:LOCALAPPDATA 'Programs\gumlet'
}

function Get-Arch {
    try {
        $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
    } catch {
        $arch = $env:PROCESSOR_ARCHITECTURE.ToLowerInvariant()
    }
    switch ($arch) {
        'x64'   { return 'x86_64' }
        'amd64' { return 'x86_64' }
        'arm64' { return 'arm64' }
        default { throw "Unsupported architecture: $arch" }
    }
}

$Arch = Get-Arch
$Archive = "gumlet_windows_${Arch}.zip"

if ($Version) {
    $Normalized = $Version.TrimStart('v')
    if ($Normalized -notmatch '^[0-9A-Za-z._-]+$') {
        throw "Invalid GUMLET_VERSION: $Version"
    }
    $BaseUrl = "https://github.com/$Repo/releases/download/v$Normalized"
    $Label = "v$Normalized"
} else {
    $BaseUrl = "https://github.com/$Repo/releases/latest/download"
    $Label = 'latest'
}

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12 -bor [Net.SecurityProtocolType]::Tls13

$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("gumlet-install-" + [System.Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

try {
    $ChecksumsPath = Join-Path $TempDir 'checksums.txt'
    $ArchivePath = Join-Path $TempDir $Archive
    $ExtractDir = Join-Path $TempDir 'extract'

    Write-Host "Downloading gumlet ($Label) for windows/$Arch..."
    Invoke-WebRequest -Uri "$BaseUrl/checksums.txt" -OutFile $ChecksumsPath -UseBasicParsing
    Invoke-WebRequest -Uri "$BaseUrl/$Archive" -OutFile $ArchivePath -UseBasicParsing

    $MatchesList = @()
    foreach ($Line in (Get-Content -Path $ChecksumsPath)) {
        if ($Line -match '^([0-9a-fA-F]{64})\s+\*?(.+)$') {
            if ($Matches[2].Trim() -eq $Archive) {
                $MatchesList += $Matches[1].ToLowerInvariant()
            }
        }
    }
    if ($MatchesList.Count -ne 1) {
        throw "Expected one SHA-256 checksum for $Archive"
    }

    $ActualHash = (Get-FileHash -Path $ArchivePath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($ActualHash -ne $MatchesList[0]) {
        throw "SHA-256 checksum mismatch for $Archive (expected $($MatchesList[0]), got $ActualHash)"
    }

    Expand-Archive -Path $ArchivePath -DestinationPath $ExtractDir -Force
    $ExtractedBinary = Join-Path $ExtractDir $Binary
    if (-not (Test-Path -Path $ExtractedBinary -PathType Leaf)) {
        throw "Archive does not contain $Binary"
    }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    $TargetPath = Join-Path $InstallDir $Binary
    Copy-Item -Path $ExtractedBinary -Destination $TargetPath -Force

    $UserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    $PathEntries = @()
    if ($UserPath) {
        $PathEntries = $UserPath -split ';' | Where-Object { $_ }
    }
    if ($PathEntries -notcontains $InstallDir) {
        $NewUserPath = if ($UserPath) { "$UserPath;$InstallDir" } else { $InstallDir }
        [Environment]::SetEnvironmentVariable('Path', $NewUserPath, 'User')
        $env:Path = "$env:Path;$InstallDir"
        Write-Host "Added $InstallDir to your user PATH."
    }

    Write-Host "Installed gumlet to $TargetPath"
    Write-Host "Run 'gumlet --version' in a new terminal to verify."
} finally {
    Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
}
