# Contributing

Thanks for taking the time to improve Controli.

## Development setup

```bash
go test ./...
make build
```

## Pull requests

Before opening a pull request:

1. Keep the change focused.
2. Add tests for behavior changes.
3. Update documentation when commands, security behavior, or files change.
4. Run `CGO_ENABLED=0 go test ./...`.

## Security changes

Controli controls access to a real machine. Treat security related changes as high risk. Be explicit about the threat model, what the change enforces, and what it does not enforce.

## Style

Use the existing Go style. Prefer small functions, typed APIs, and plain error messages. Do not add network services, background daemons, or automatic system changes without a clear review path.
