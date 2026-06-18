# Compatibility

Controli is built as static Go binaries with `CGO_ENABLED=0`. That keeps installation simple and avoids runtime library problems on most machines.

## Support Model

Controli targets operating systems and CPU architectures supported by the Go toolchain used for the release. It cannot guarantee every historical OS version, every vendor image, or damaged system installs.

The intended support goal is:

- Current and recent Windows versions for joining sessions and best-effort hosting.
- Intel and Apple Silicon Macs for hosting and joining.
- Common Ubuntu and Linux systems for hosting and joining.
- Additional Linux server and board architectures where Go can build the binary.

## Host And Guest Support

| Platform | Join | Host |
| --- | --- | --- |
| macOS Intel | Yes | Yes |
| macOS Apple Silicon | Yes | Yes |
| Linux amd64 | Yes | Yes |
| Linux arm64 | Yes | Yes |
| Linux 386 | Yes | Best effort |
| Linux armv6 | Yes | Best effort |
| Linux armv7 | Yes | Best effort |
| Linux ppc64le | Yes | Best effort |
| Linux riscv64 | Yes | Best effort |
| Linux s390x | Yes | Best effort |
| Windows amd64 | Yes | Best effort |
| Windows 386 | Yes | Best effort |
| Windows arm | Yes | Best effort |
| Windows arm64 | Yes | Best effort |

macOS and Linux hosting use a real PTY and can keep shells alive through `tmux` when it is installed. Windows hosting uses a stdio backend until ConPTY is added.

## Release Assets

| Asset | Use |
| --- | --- |
| `controli-darwin-arm64` | Apple Silicon Macs |
| `controli-darwin-amd64` | Intel Macs |
| `controli-linux-386` | 32-bit Intel or AMD Linux |
| `controli-linux-amd64` | Common Ubuntu, Debian, Fedora, and server Linux |
| `controli-linux-armv6` | ARMv6 boards |
| `controli-linux-armv7` | ARMv7 boards |
| `controli-linux-arm64` | ARM64 Linux servers and boards |
| `controli-linux-ppc64le` | PowerPC 64 little-endian Linux |
| `controli-linux-riscv64` | RISC-V 64 Linux |
| `controli-linux-s390x` | IBM Z and LinuxONE |
| `controli-windows-386.exe` | 32-bit Windows |
| `controli-windows-amd64.exe` | Common Intel or AMD Windows |
| `controli-windows-arm.exe` | 32-bit Windows ARM |
| `controli-windows-arm64.exe` | Windows on ARM64 |

## Practical Limits

- Very old macOS releases may not run current Go-built binaries.
- Very old Linux kernels may not run current Go-built binaries.
- Host mode requires a working shell.
- macOS and Linux host mode requires PTY support.
- Persistent macOS and Linux host sessions require `tmux`.
- Tunnel join mode requires outbound HTTPS and WebSocket access to the public tunnel hostname.
- Relay fallback mode requires outbound HTTPS and WebSocket access to the relay.
