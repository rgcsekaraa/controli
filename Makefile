.PHONY: test build docs clean

GO ?= go
GOFLAGS ?=
DIST := dist

TARGETS := \
	controli-darwin-arm64 \
	controli-darwin-amd64 \
	controli-linux-386 \
	controli-linux-amd64 \
	controli-linux-armv6 \
	controli-linux-armv7 \
	controli-linux-arm64 \
	controli-linux-ppc64le \
	controli-linux-riscv64 \
	controli-linux-s390x \
	controli-windows-386.exe \
	controli-windows-amd64.exe \
	controli-windows-arm.exe \
	controli-windows-arm64.exe

test:
	CGO_ENABLED=0 $(GO) test ./...

docs:
	npm run docs:build

build:
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(DIST)/controli-darwin-arm64 ./cmd/controli
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(DIST)/controli-darwin-amd64 ./cmd/controli
	CGO_ENABLED=0 GOOS=linux GOARCH=386 $(GO) build $(GOFLAGS) -o $(DIST)/controli-linux-386 ./cmd/controli
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(DIST)/controli-linux-amd64 ./cmd/controli
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 $(GO) build $(GOFLAGS) -o $(DIST)/controli-linux-armv6 ./cmd/controli
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 $(GO) build $(GOFLAGS) -o $(DIST)/controli-linux-armv7 ./cmd/controli
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(DIST)/controli-linux-arm64 ./cmd/controli
	CGO_ENABLED=0 GOOS=linux GOARCH=ppc64le $(GO) build $(GOFLAGS) -o $(DIST)/controli-linux-ppc64le ./cmd/controli
	CGO_ENABLED=0 GOOS=linux GOARCH=riscv64 $(GO) build $(GOFLAGS) -o $(DIST)/controli-linux-riscv64 ./cmd/controli
	CGO_ENABLED=0 GOOS=linux GOARCH=s390x $(GO) build $(GOFLAGS) -o $(DIST)/controli-linux-s390x ./cmd/controli
	CGO_ENABLED=0 GOOS=windows GOARCH=386 $(GO) build $(GOFLAGS) -o $(DIST)/controli-windows-386.exe ./cmd/controli
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(DIST)/controli-windows-amd64.exe ./cmd/controli
	CGO_ENABLED=0 GOOS=windows GOARCH=arm $(GO) build $(GOFLAGS) -o $(DIST)/controli-windows-arm.exe ./cmd/controli
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(DIST)/controli-windows-arm64.exe ./cmd/controli

clean:
	rm -f $(addprefix $(DIST)/,$(TARGETS))
