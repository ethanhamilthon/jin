GO ?= go
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
BINARY := bin/jin
BUILD_MODE := -ldflags "-X jin/internal/paths.mode=prod"

.PHONY: build build-prod run test vet check golden install release clean web web-dev web-check

build: web
	$(GO) build -o "$(BINARY)" .

build-prod: web
	$(GO) build $(BUILD_MODE) -o "$(BINARY)" .

# web builds the browser UI of jin web into internal/web/dist. Without Node
# the Go build still works, and jin web says that the UI is missing.
web:
	@if command -v npm >/dev/null 2>&1; then \
		cd web && npm ci --no-audit --no-fund && npm run build; \
	else \
		echo "npm not found: building without the jin web UI"; \
	fi

# web-dev serves the UI with hot reload; run jin web --port 7373 next to it
# and open the Vite address with ?token=<token from jin web>.
web-dev:
	cd web && npm run dev

web-check:
	cd web && npm ci --no-audit --no-fund && npm run check

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
	find internal/web/dist -mindepth 1 ! -name .gitkeep -exec rm -rf {} +
