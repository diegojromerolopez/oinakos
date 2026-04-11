.PHONY: all build build-wasm build-tools test run run-debug boundaries-editor map-editor serve-wasm bundle-mac bundle-windows bundle-linux bundle-all release clean

# Default name for the native binary
APP_NAME=oinakos

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean

# Output directories
BIN_DIR=bin
DIST_DIR=dist

all: build build-wasm build-tools dist

# Versioning
VERSION=0.2-alpha
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"

build:
	@echo "Building native binary $(VERSION)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) .
	@echo "Built: $(BIN_DIR)/$(APP_NAME)"

build-wasm:
	@echo "Building WebAssembly binary $(VERSION)..."
	@mkdir -p $(DIST_DIR)
	GOOS=js GOARCH=wasm $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME).wasm .
	@echo "Built: $(DIST_DIR)/$(APP_NAME).wasm"

build-tools: $(BIN_DIR)/boundaries_editor $(BIN_DIR)/map_editor

$(BIN_DIR)/boundaries_editor: ./tools/boundaries_editor/*.go
	@echo "Building boundaries_editor..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/boundaries_editor ./tools/boundaries_editor/
	@echo "Tool built: $(BIN_DIR)/boundaries_editor"

$(BIN_DIR)/map_editor: ./tools/map_editor/*.go
	@echo "Building map_editor..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/map_editor ./tools/map_editor/
	@echo "Tool built: $(BIN_DIR)/map_editor"

test:
	@echo "Running tests..."
	$(GOCMD) test -tags test ./internal/...

# Portable sed -i
ifeq ($(shell uname), Darwin)
  SED_I = sed -i ''
else
  SED_I = sed -i
endif

dist: build-wasm
	@echo "Preparing distribution files..."
	@mkdir -p $(DIST_DIR)
	@cp assets/images/logo.png $(DIST_DIR)/
	@# Copy wasm_exec.js for reference, but we will also inline it
	@cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(DIST_DIR)/
	@# Use the template from assets/wasm/index.html
	@cp assets/wasm/index.html $(DIST_DIR)/index.html
	@# Inline wasm_exec.js into index.html
	@$(SED_I) '/\/\*WASM_EXEC_JS\*\//r $(DIST_DIR)/wasm_exec.js' $(DIST_DIR)/index.html
	@# Inject version
	@$(SED_I) 's/{{VERSION}}/$(VERSION)/g' $(DIST_DIR)/index.html
	@rm $(DIST_DIR)/wasm_exec.js
	@echo "Dist files prepared in $(DIST_DIR)/: index.html (inlined JS), logo.png and oinakos.wasm"



run: build
	./$(BIN_DIR)/$(APP_NAME) $(if $(MAP),-map $(MAP)) $(if $(MAP_TYPE),-map-type $(MAP_TYPE)) $(if $(HERO),-hero $(HERO))

run-debug: build
	./$(BIN_DIR)/$(APP_NAME) -debug $(if $(MAP),-map $(MAP)) $(if $(MAP_TYPE),-map-type $(MAP_TYPE)) $(if $(HERO),-hero $(HERO))
run-headless:
	@echo "Running Headless Simulation..."
	$(GORUN) -tags headless .

boundaries-editor: build-tools
	@if [ -z "$(OBSTACLE)$(NPC)$(CHARACTER)$(OBJECT)" ]; then ./$(BIN_DIR)/boundaries_editor; else \
	./$(BIN_DIR)/boundaries_editor \
		$(if $(OBSTACLE),--obstacle $(OBSTACLE)) \
		$(if $(NPC),--npc $(NPC)) \
		$(if $(CHARACTER),--character $(CHARACTER)) \
		$(if $(OBJECT),--object $(OBJECT)); \
	fi

map-editor: build-tools
	./$(BIN_DIR)/map_editor

serve-wasm: dist
	@echo "Serving WASM on port 8000..."
	@cd $(DIST_DIR) && python3 -m http.server 8000

bundle-mac:
	@echo "Bundling for macOS $(VERSION)..."
	@chmod +x scripts/bundle_mac.sh
	@VERSION=$(VERSION) ./scripts/bundle_mac.sh

bundle-windows:
	@echo "Bundling for Windows $(VERSION)..."
	@chmod +x scripts/bundle_windows.sh
	@VERSION=$(VERSION) ./scripts/bundle_windows.sh

bundle-linux:
	@echo "Bundling for Linux $(VERSION)..."
	@chmod +x scripts/bundle_linux.sh
	@VERSION=$(VERSION) ./scripts/bundle_linux.sh

bundle-all: bundle-mac bundle-windows bundle-linux
	@echo "All platforms bundled successfully."

release: test
	@echo "Releasing version $(VERSION)..."
	@if git rev-parse $(VERSION) >/dev/null 2>&1; then echo "Error: Tag $(VERSION) already exists."; exit 1; fi
	@git diff-index --quiet HEAD -- || (echo "Error: Uncommitted changes in repository. Please commit or stash them before releasing." && exit 1)
	git tag -a $(VERSION) -m "Release version $(VERSION)"
	git push origin $(VERSION)
	@echo "Version $(VERSION) released and pushed successfully."

clean:
	@echo "Cleaning up..."
	$(GOCLEAN)
	rm -rf $(BIN_DIR)
	rm -rf $(DIST_DIR)
	rm -rf tools/bin
	@echo "Cleaned."
