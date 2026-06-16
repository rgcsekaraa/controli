# Install

Use the binary that matches the target machine.

## Windows

| Machine | Binary |
| --- | --- |
| Most Intel or AMD PCs | `controli-windows-amd64.exe` |
| Older 32-bit PCs | `controli-windows-386.exe` |
| Windows on ARM64 | `controli-windows-arm64.exe` |
| Older Windows ARM devices | `controli-windows-arm.exe` |

Example:

```powershell
Invoke-WebRequest -Uri "https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-windows-amd64.exe" -OutFile "$env:USERPROFILE\Downloads\controli.exe"
& "$env:USERPROFILE\Downloads\controli.exe" join 1234567
```

## macOS

| Machine | Binary |
| --- | --- |
| Apple Silicon | `controli-darwin-arm64` |
| Intel Mac | `controli-darwin-amd64` |

Example:

```bash
curl -L -o controli https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-darwin-arm64
chmod +x controli
./controli join 1234567
```

## Linux and Ubuntu

| Machine | Binary |
| --- | --- |
| Most Intel or AMD desktops and servers | `controli-linux-amd64` |
| Older 32-bit Intel or AMD systems | `controli-linux-386` |
| ARMv6 boards | `controli-linux-armv6` |
| ARMv7 boards | `controli-linux-armv7` |
| ARM64 servers and boards | `controli-linux-arm64` |
| PowerPC 64 little-endian servers | `controli-linux-ppc64le` |
| RISC-V 64 systems | `controli-linux-riscv64` |
| IBM Z or LinuxONE | `controli-linux-s390x` |

Example:

```bash
curl -L -o controli https://github.com/rgcsekaraa/controli/releases/download/v0.2.0/controli-linux-amd64
chmod +x controli
./controli join 1234567
```

After downloading on Unix systems, mark the file executable with `chmod +x`.
