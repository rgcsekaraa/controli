# Changelog

All notable changes to Controli will be documented in this file.

## Unreleased

- Added enforced Authenticode signing for tagged Windows release builds.
- Added SHA256 checksum generation for release assets.
- Documented the Azure Trusted Signing setup required for Windows Device Guard and Smart App Control.

## 0.4.2

- Enforced one active guest connection per invite code.
- Allowed the same guest to rejoin with the same valid code after disconnect.
- Reset host approval on every guest connection and reconnection.
- Rejected additional active guests instead of allowing shared full-control input.
- Updated relay health metadata and documentation for the single-guest security model.

## 0.4.1

- Allowed multiple guests to join the same live session with the same 7-digit code until the invite expires.
- Added per-client relay socket IDs so relay fallback no longer replaces the first guest when another guest joins.
- Broadcast host terminal output to all connected relay guests.
- Documented same-code multi-guest behavior.

## 0.4.0

- Added tunnel mode for long sessions with `controli host tunnel`.
- Added direct browser terminal serving from the host over a named Cloudflare Tunnel.
- Moved 7-digit invite storage from Durable Objects to Workers KV.
- Kept Durable Object relay mode as fallback through `controli host share`.
- Added tunnel mode documentation, command reference updates, and long-session troubleshooting.
- Updated the browser terminal to use `wss://` automatically on HTTPS pages.

## 0.3.0

- Added host approval prompt.
- Added permission modes: `full`, `view`, and `approve`.
- Added audit logs and session status output.
- Added Windows hosting through a stdio backend.
- Added byte-bounded relay queues and stronger reconnect behavior.
- Added `controli update`, install scripts, and expanded command documentation.

## 0.2.1-go-alpha

- Added Go core CLI.
- Added macOS/Linux PTY hosting.
- Added cross-platform relay joining.
- Added embedded xterm.js browser terminal for reliable rendering on Windows, macOS, and Linux.
- Added Cloudflare Worker short-code invite compatibility.
- Added pure-Go release binaries for macOS, Linux, and Windows.

## 0.1.0

- Initial prototype with SSH/session-management experiments.
