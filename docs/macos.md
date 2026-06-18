# macOS

macOS can host and join sessions.

Apple Silicon uses the `darwin-arm64` binary. Intel Macs use the `darwin-amd64` binary.

When hosting, Controli starts the configured shell in a PTY. Install `tmux` to keep hosted shells alive across Controli restarts and network drops.

```bash
brew install tmux
```
