.PHONY: all build build-wasm build-tools run run-debug view-boundaries serve-wasm bundle-mac bundle-windows bundle-linux bundle-all clean

# Default name for the native binary
APP_NAME=oinakos

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean

# Output directory
BIN_DIR=bin

all: build build-wasm build-tools

build:
	@echo "Building native binary..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(APP_NAME) main.go
	@echo "Built: $(BIN_DIR)/$(APP_NAME)"

build-wasm:
	@echo "Building WebAssembly binary..."
	@mkdir -p $(BIN_DIR)
	GOOS=js GOARCH=wasm $(GOBUILD) -o $(BIN_DIR)/$(APP_NAME).wasm main.go
	@echo "Copying wasm_exec.js..."
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(BIN_DIR)/
	@echo "Built: $(BIN_DIR)/$(APP_NAME).wasm"

build-tools: tools/bin/view_boundaries

tools/bin/view_boundaries: ./tools/view_boundaries/main.go
	@echo "Building view_boundaries..."
	@mkdir -p tools/bin
	$(GOBUILD) -o tools/bin/view_boundaries ./tools/view_boundaries/main.go
	@echo "Tool built: tools/bin/view_boundaries"

run: build
	./$(BIN_DIR)/$(APP_NAME)

run-debug: build
	./$(BIN_DIR)/$(APP_NAME) -debug

view-boundaries: tools/bin/view_boundaries
	@if [ -z "$(OBSTACLE)$(NPC)$(CHARACTER)" ]; then echo "Usage: make view-boundaries [OBSTACLE=id | NPC=id | CHARACTER=main]"; exit 1; fi
	./tools/bin/view_boundaries \
		$(if $(OBSTACLE),--obstacle $(OBSTACLE)) \
		$(if $(NPC),--npc $(NPC)) \
		$(if $(CHARACTER),--character $(CHARACTER))

serve-wasm: build-wasm
	@echo "Serving WASM on port 8000..."
	@cd $(BIN_DIR) && python3 -m http.server 8000

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
	rm -rf tools/bin
	@echo "Cleaned."
