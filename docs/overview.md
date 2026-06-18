# Overview

Controli shares a terminal session through outbound connections.

The host starts a shell in a PTY. The guest joins with a short code. For long sessions, terminal traffic flows through a named Cloudflare Tunnel directly to the host terminal server.

The Cloudflare Worker stores short invite codes in Workers KV. The Durable Object relay remains available only as a fallback transport.

The default guest renderer is a local browser terminal backed by embedded xterm.js assets. This keeps rendering consistent across Windows, macOS, and Linux.
