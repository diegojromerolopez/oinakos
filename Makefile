.PHONY: all build build-wasm run serve-wasm clean

# Default name for the native binary
APP_NAME=oinakos

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean

# Output directory
BIN_DIR=bin

all: build build-wasm

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

run: build
	@echo "Running native game..."
	./$(BIN_DIR)/$(APP_NAME)

serve-wasm: build-wasm
	@echo "Serving WASM on port 8000..."
	@cd $(BIN_DIR) && python3 -m http.server 8000

clean:
	@echo "Cleaning up..."
	$(GOCLEAN)
	rm -rf $(BIN_DIR)
	@echo "Cleaned."
