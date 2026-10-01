GO ?= go
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
BINARY := bin/jin
BUILD_MODE := -ldflags "-X jin/internal/paths.mode=prod"

.PHONY: build build-prod run test vet check install release clean

build:
	$(GO) build -o "$(BINARY)" .

build-prod:
	$(GO) build $(BUILD_MODE) -o "$(BINARY)" .

run:
	$(GO) run .

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

check: test vet

install: build-prod
	install -d "$(BINDIR)"
	install -m 755 "$(BINARY)" "$(BINDIR)/jin"

release:
	scripts/build-release.sh "$(VERSION)"

clean:
	rm -rf "$(BINARY)" dist
