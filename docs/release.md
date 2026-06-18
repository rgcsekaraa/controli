# Release

Release assets are built for:

- macOS arm64
- macOS amd64
- Linux 386
- Linux amd64
- Linux armv6
- Linux armv7
- Linux arm64
- Linux ppc64le
- Linux riscv64
- Linux s390x
- Windows 386
- Windows amd64
- Windows arm
- Windows arm64

The GitHub release workflow uses the same target names as the local `make build` command.

## Windows Trust

Windows `.exe` assets are currently unsigned. On personal Windows machines this usually runs after the user allows the downloaded file. On machines managed with Device Guard, App Control for Business, Smart App Control, or WDAC, each new release hash may need to be allowed by policy.

The release job publishes `SHA256SUMS.txt` for all assets so administrators can allow a specific release hash when publisher signing is not available.
