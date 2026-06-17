$ErrorActionPreference = "Stop"

$Repo = if ($env:CONTROLI_REPO) { $env:CONTROLI_REPO } else { "rgcsekaraa/controli" }
$InstallDir = if ($env:CONTROLI_INSTALL_DIR) { $env:CONTROLI_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "Controli" }
$Arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()

switch ($Arch) {
  "x64" { $Asset = "controli-windows-amd64.exe" }
  "x86" { $Asset = "controli-windows-386.exe" }
  "arm" { $Asset = "controli-windows-arm.exe" }
  "arm64" { $Asset = "controli-windows-arm64.exe" }
  default { throw "unsupported architecture: $Arch" }
}

$Url = "https://github.com/$Repo/releases/latest/download/$Asset"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$Target = Join-Path $InstallDir "controli.exe"
Invoke-WebRequest -Uri $Url -OutFile $Target
Write-Host "installed $Target"
Write-Host "run: `"$Target`" join <code>"
