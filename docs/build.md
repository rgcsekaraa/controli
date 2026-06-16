# Build

Run tests:

```bash
CGO_ENABLED=0 go test ./...
```

Build release binaries:

```bash
make build
```

The release build uses pure Go binaries for predictable cross-platform packaging.

