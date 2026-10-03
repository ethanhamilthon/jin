GO ?= go
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
BINARY := bin/jin
BUILD_MODE := -ldflags "-X jin/internal/paths.mode=prod"

.PHONY: build build-prod run test vet check golden install release clean

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

# golden rewrites the TUI screen snapshots in internal/ui/testdata after an
# intended change to the look; review the diff before committing.
golden:
	$(GO) test ./internal/ui -run Golden -update

install: build-prod
	install -d "$(BINDIR)"
	install -m 755 "$(BINARY)" "$(BINDIR)/jin"

release:
	scripts/build-release.sh "$(VERSION)"

clean:
	rm -rf "$(BINARY)" dist
