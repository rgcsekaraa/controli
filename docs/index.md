---
layout: home

title: Controli
titleTemplate: Native Go CLI sharing

hero:
  name: Controli
  text: Native CLI sharing for support sessions
  tagline: Share a real terminal through an outbound Cloudflare relay with a short invite code and a browser-based guest terminal.
  image:
    src: /logo.svg
    alt: Controli
  actions:
    - theme: brand
      text: Get Started
      link: /install
    - theme: alt
      text: View on GitHub
      link: https://github.com/rgcsekaraa/controli

features:
  - title: Simple session codes
    details: Start a host share, send a 7-digit code, and let the guest join without inbound ports.
  - title: Browser terminal renderer
    details: Guests use an embedded xterm.js terminal for consistent behavior across Windows, macOS, and Linux.
  - title: Native Go binaries
    details: Release assets are built for common desktop and server platforms with no runtime installer required.
  - title: Cloudflare relay
    details: The bundled Worker handles rendezvous, short-code lookup, and WebSocket forwarding.
---

## Quick Start

Configure the relay:

```bash
controli relay configure --url wss://controli-relay.example.workers.dev
```

Start a host session:

```bash
controli host share --workspace main --minutes 480
```

Join from another machine:

```bash
controli join 1234567
```

## What To Read Next

- [Install](install.md)
- [Host](host.md)
- [Join](join.md)
- [Security](security.md)
