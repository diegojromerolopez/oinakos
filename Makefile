.PHONY: all build build-wasm build-tools run run-debug boundaries-editor serve-wasm bundle-mac bundle-windows bundle-linux bundle-all clean

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

build-tools: tools/bin/boundaries_editor

tools/bin/boundaries_editor: ./tools/boundaries_editor/main.go
	@echo "Building boundaries_editor..."
	@mkdir -p tools/bin
	$(GOBUILD) -o tools/bin/boundaries_editor ./tools/boundaries_editor/main.go
	@echo "Tool built: tools/bin/boundaries_editor"

run: build
	./$(BIN_DIR)/$(APP_NAME)

run-debug: build
	./$(BIN_DIR)/$(APP_NAME) -debug

boundaries-editor: tools/bin/boundaries_editor
	@if [ -z "$(OBSTACLE)$(NPC)$(CHARACTER)" ]; then ./tools/bin/boundaries_editor; else \
	./tools/bin/boundaries_editor \
		$(if $(OBSTACLE),--obstacle $(OBSTACLE)) \
		$(if $(NPC),--npc $(NPC)) \
		$(if $(CHARACTER),--character $(CHARACTER)); \
	fi

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
