param(
    [string]$Version = "",
    [string]$InstallDir = "$env:LOCALAPPDATA\easy8\bin"
)

$ErrorActionPreference = "Stop"

$Repo = "Easy8Com/easy8-cli"

$archRaw = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
switch ($archRaw) {
    "x64"   { $Arch = "amd64" }
    "arm64" { $Arch = "arm64" }
    default  { throw "Unsupported architecture: $archRaw" }
}

if ([string]::IsNullOrWhiteSpace($Version)) {
    Write-Host "Fetching latest version..."
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    $Version = $release.tag_name
}

if ([string]::IsNullOrWhiteSpace($Version)) {
    throw "Failed to determine release version. Make sure a GitHub release exists."
}

$assetName = "easy8-windows-$Arch.exe"
$binaryUrl = "https://github.com/$Repo/releases/download/$Version/$assetName"
$checksumsUrl = "https://github.com/$Repo/releases/download/$Version/checksums.txt"

$tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ("easy8-install-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -Path $tmpDir -ItemType Directory | Out-Null

try {
    $binaryPath = Join-Path $tmpDir $assetName
    $checksumsPath = Join-Path $tmpDir "checksums.txt"

    Write-Host "Downloading $assetName ($Version)..."
    Invoke-WebRequest -Uri $binaryUrl -OutFile $binaryPath
    Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsPath

    $expected = $null
    foreach ($line in Get-Content -Path $checksumsPath) {
        if ($line -match "\s+$([regex]::Escape($assetName))$") {
            $expected = ($line -split '\s+')[0].ToLowerInvariant()
            break
        }
    }

    if ([string]::IsNullOrWhiteSpace($expected)) {
        throw "Checksum entry for $assetName was not found in checksums.txt"
    }

    $actual = (Get-FileHash -Path $binaryPath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($expected -ne $actual) {
        throw "Checksum mismatch for $assetName. Expected $expected, got $actual"
    }

    New-Item -Path $InstallDir -ItemType Directory -Force | Out-Null
    $targetPath = Join-Path $InstallDir "easy8.exe"
    Copy-Item -Path $binaryPath -Destination $targetPath -Force

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($null -eq $userPath) {
        $userPath = ""
    }

    $pathEntries = @()
    if (-not [string]::IsNullOrWhiteSpace($userPath)) {
        $pathEntries = $userPath.Split(';') | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
    }

    if (-not ($pathEntries -contains $InstallDir)) {
        $newUserPath = if ([string]::IsNullOrWhiteSpace($userPath)) { $InstallDir } else { "$InstallDir;$userPath" }
        [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")

        if ([string]::IsNullOrWhiteSpace($env:Path)) {
            $env:Path = $InstallDir
        } else {
            $env:Path = "$InstallDir;$env:Path"
        }

        Write-Host "Added $InstallDir to your user PATH."
        Write-Host "If 'easy8' is still not found, restart your terminal."
    }

    Write-Host ""
    Write-Host "easy8 $Version installed to $targetPath"
    & $targetPath version
    Write-Host "Run 'easy8 setup' to configure Easy8 API access."
}
finally {
    Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
}
