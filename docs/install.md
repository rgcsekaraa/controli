# Install

Download the binary that matches the guest or host machine. The current release is [v0.4.2](https://github.com/rgcsekaraa/controli/releases/tag/v0.4.2).

<div class="download-grid">
  <section class="download-card">
    <h2>Windows</h2>
    <p>Use these when the other person is joining from a Windows machine.</p>
    <table>
      <thead>
        <tr><th>Machine</th><th>Download</th></tr>
      </thead>
      <tbody>
        <tr><td>Most Intel or AMD PCs</td><td><a href="https://github.com/rgcsekaraa/controli/releases/download/v0.4.2/controli-windows-amd64.exe">controli-windows-amd64.exe</a></td></tr>
        <tr><td>Older 32-bit PCs</td><td><a href="https://github.com/rgcsekaraa/controli/releases/download/v0.4.2/controli-windows-386.exe">controli-windows-386.exe</a></td></tr>
        <tr><td>Windows on ARM64</td><td><a href="https://github.com/rgcsekaraa/controli/releases/download/v0.4.2/controli-windows-arm64.exe">controli-windows-arm64.exe</a></td></tr>
        <tr><td>Older Windows ARM devices</td><td><a href="https://github.com/rgcsekaraa/controli/releases/download/v0.4.2/controli-windows-arm.exe">controli-windows-arm.exe</a></td></tr>
      </tbody>
    </table>
  </section>

  <section class="download-card">
    <h2>macOS</h2>
    <p>Use these for Apple machines. Apple Silicon means M1, M2, M3, or newer.</p>
    <table>
      <thead>
        <tr><th>Machine</th><th>Download</th></tr>
      </thead>
      <tbody>
        <tr><td>Apple Silicon</td><td><a href="https://github.com/rgcsekaraa/controli/releases/download/v0.4.2/controli-darwin-arm64">controli-darwin-arm64</a></td></tr>
        <tr><td>Intel Mac</td><td><a href="https://github.com/rgcsekaraa/controli/releases/download/v0.4.2/controli-darwin-amd64">controli-darwin-amd64</a></td></tr>
      </tbody>
    </table>
  </section>

  <section class="download-card">
    <h2>Linux and Ubuntu</h2>
    <p>Use these for Ubuntu, Debian, Fedora, servers, boards, and other Linux systems.</p>
    <table>
      <thead>
        <tr><th>Machine</th><th>Download</th></tr>
      </thead>
      <tbody>
        <tr><td>Most Intel or AMD desktops and servers</td><td><a href="https://github.com/rgcsekaraa/controli/releases/download/v0.4.2/controli-linux-amd64">controli-linux-amd64</a></td></tr>
        <tr><td>Older 32-bit Intel or AMD systems</td><td><a href="https://github.com/rgcsekaraa/controli/releases/download/v0.4.2/controli-linux-386">controli-linux-386</a></td></tr>
        <tr><td>ARM64 servers and boards</td><td><a href="https://github.com/rgcsekaraa/controli/releases/download/v0.4.2/controli-linux-arm64">controli-linux-arm64</a></td></tr>
        <tr><td>ARMv7 boards</td><td><a href="https://github.com/rgcsekaraa/controli/releases/download/v0.4.2/controli-linux-armv7">controli-linux-armv7</a></td></tr>
        <tr><td>ARMv6 boards</td><td><a href="https://github.com/rgcsekaraa/controli/releases/download/v0.4.2/controli-linux-armv6">controli-linux-armv6</a></td></tr>
        <tr><td>PowerPC 64 little-endian servers</td><td><a href="https://github.com/rgcsekaraa/controli/releases/download/v0.4.2/controli-linux-ppc64le">controli-linux-ppc64le</a></td></tr>
        <tr><td>RISC-V 64 systems</td><td><a href="https://github.com/rgcsekaraa/controli/releases/download/v0.4.2/controli-linux-riscv64">controli-linux-riscv64</a></td></tr>
        <tr><td>IBM Z or LinuxONE</td><td><a href="https://github.com/rgcsekaraa/controli/releases/download/v0.4.2/controli-linux-s390x">controli-linux-s390x</a></td></tr>
      </tbody>
    </table>
  </section>
</div>

## Quick Download

Windows, most PCs:

```powershell
Invoke-WebRequest -Uri "https://github.com/rgcsekaraa/controli/releases/latest/download/controli-windows-amd64.exe" -OutFile "$env:USERPROFILE\Downloads\controli.exe"
& "$env:USERPROFILE\Downloads\controli.exe" join 1234567
```

Windows release binaries are currently unsigned. If Device Guard or App Control blocks the file on a company-managed machine, the organization's policy must allow the release hash.

macOS, Apple Silicon:

```bash
curl -L -o controli https://github.com/rgcsekaraa/controli/releases/latest/download/controli-darwin-arm64
chmod +x controli
./controli join 1234567
```

Linux or Ubuntu, most PCs and servers:

```bash
curl -L -o controli https://github.com/rgcsekaraa/controli/releases/latest/download/controli-linux-amd64
chmod +x controli
./controli join 1234567
```

After downloading on Unix systems, mark the file executable with `chmod +x`.

## One-command install

macOS and Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/rgcsekaraa/controli/main/scripts/install.sh | sh
```

Windows:

```powershell
iwr https://raw.githubusercontent.com/rgcsekaraa/controli/main/scripts/install.ps1 -UseB | iex
```

Update later:

```bash
controli update
```
