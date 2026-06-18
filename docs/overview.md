# Overview

Controli shares a terminal session through outbound connections.

The host starts a shell in a PTY. The guest joins with a short code from the hosted browser join page or the CLI. For long sessions, terminal traffic flows through a named Cloudflare Tunnel directly to the host terminal server.

The Cloudflare Worker stores short invite codes in Workers KV. The Durable Object relay remains available only as a fallback transport.

Tunnel sessions can be joined from a normal browser without installing Controli on the guest machine. CLI join remains available for relay fallback sessions and diagnostics.
