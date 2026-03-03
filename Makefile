.PHONY: all build build-wasm build-tools test run run-debug boundaries-editor map-editor serve-wasm bundle-mac bundle-windows bundle-linux bundle-all clean

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

build:
	@echo "Building native binary..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(APP_NAME) main.go
	@echo "Built: $(BIN_DIR)/$(APP_NAME)"

build-wasm:
	@echo "Building WebAssembly binary..."
	@mkdir -p $(DIST_DIR)
	GOOS=js GOARCH=wasm $(GOBUILD) -o $(DIST_DIR)/$(APP_NAME).wasm main.go
	@echo "Built: $(DIST_DIR)/$(APP_NAME).wasm"

build-tools: $(BIN_DIR)/boundaries_editor $(BIN_DIR)/map_editor

$(BIN_DIR)/boundaries_editor: ./tools/boundaries_editor/main.go
	@echo "Building boundaries_editor..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/boundaries_editor ./tools/boundaries_editor/main.go
	@echo "Tool built: $(BIN_DIR)/boundaries_editor"

$(BIN_DIR)/map_editor: ./tools/map_editor/main.go
	@echo "Building map_editor..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/map_editor ./tools/map_editor/main.go
	@echo "Tool built: $(BIN_DIR)/map_editor"

test:
	@echo "Running tests..."
	$(GOCMD) test ./...

dist: build-wasm
	@echo "Preparing distribution files..."
	@mkdir -p $(DIST_DIR)
	@# Copy wasm_exec.js for reference, but we will also inline it
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(DIST_DIR)/
	@# Generate index.html with inlined wasm_exec.js and glue code
	@echo "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>Oinakos</title><style>body { margin: 0; background: #000; overflow: hidden; display: flex; justify-content: center; align-items: center; height: 100vh; font-family: sans-serif; }</style></head><body><div id=\"status\" style=\"color: #daa520; font-size: 24px;\">Loading Oinakos...</div><script>" > $(DIST_DIR)/index.html
	@cat $(DIST_DIR)/wasm_exec.js >> $(DIST_DIR)/index.html
	@echo "</script><script>const go = new Go(); WebAssembly.instantiateStreaming(fetch('oinakos.wasm'), go.importObject).then((result) => { document.getElementById('status').style.display = 'none'; go.run(result.instance); });</script></body></html>" >> $(DIST_DIR)/index.html
	rm $(DIST_DIR)/wasm_exec.js
	@echo "Dist files prepared in $(DIST_DIR)/: index.html (inlined JS) and oinakos.wasm"

run: build
	./$(BIN_DIR)/$(APP_NAME)

run-debug: build
	./$(BIN_DIR)/$(APP_NAME) -debug

boundaries-editor: build-tools
	@if [ -z "$(OBSTACLE)$(NPC)$(CHARACTER)" ]; then ./$(BIN_DIR)/boundaries_editor; else \
	./$(BIN_DIR)/boundaries_editor \
		$(if $(OBSTACLE),--obstacle $(OBSTACLE)) \
		$(if $(NPC),--npc $(NPC)) \
		$(if $(CHARACTER),--character $(CHARACTER)); \
	fi

map-editor: build-tools
	./$(BIN_DIR)/map_editor

serve-wasm: dist
	@echo "Serving WASM on port 8000..."
	@cd $(DIST_DIR) && python3 -m http.server 8000

bundle-mac:
	@echo "Bundling for macOS..."
	@chmod +x scripts/bundle_mac.sh
	@./scripts/bundle_mac.sh

bundle-windows:
	@echo "Bundling for Windows..."
	@chmod +x scripts/bundle_windows.sh
	@./scripts/bundle_windows.sh

bundle-linux:
	@echo "Bundling for Linux..."
	@chmod +x scripts/bundle_linux.sh
	@./scripts/bundle_linux.sh

bundle-all: bundle-mac bundle-windows bundle-linux
	@echo "All platforms bundled successfully."

clean:
	@echo "Cleaning up..."
	$(GOCLEAN)
	rm -rf $(BIN_DIR)
	rm -rf $(DIST_DIR)
	rm -rf tools/bin
	@echo "Cleaned."
