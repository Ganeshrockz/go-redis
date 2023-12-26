SHELL = /usr/bin/env bash -euo pipefail -c

BIN_NAME = go-redis
ARCH     = $(shell A=$$(uname -m); [ $$A = x86_64 ] && A=amd64; echo $$A)
OS       = $(shell uname | tr [[:upper:]] [[:lower:]])
PLATFORM = $(OS)/$(ARCH)
DIST     = dist/$(PLATFORM)
BIN      = $(DIST)/$(BIN_NAME)

dist:
	mkdir -p $(DIST)

dev: dist
	GOARCH=$(ARCH) GOOS=$(OS) go build -ldflags "$(LD_FLAGS)" -o $(BIN)
.PHONY: dev