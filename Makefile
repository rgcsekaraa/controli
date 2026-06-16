.PHONY: test build docs clean

GO ?= go
GOFLAGS ?=

test:
	CGO_ENABLED=0 $(GO) test ./...

docs:
	npm run docs:build

build:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o dist/controli-darwin-arm64 ./cmd/controli
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o dist/controli-darwin-amd64 ./cmd/controli
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o dist/controli-linux-amd64 ./cmd/controli
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -o dist/controli-linux-arm64 ./cmd/controli
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o dist/controli-windows-amd64.exe ./cmd/controli

clean:
	rm -f dist/controli-darwin-arm64 dist/controli-darwin-amd64 dist/controli-linux-amd64 dist/controli-linux-arm64 dist/controli-windows-amd64.exe
