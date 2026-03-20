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

# Versioning
VERSION=0.1-alpha
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"

build:
	@echo "Building native binary $(VERSION)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) main.go
	@echo "Built: $(BIN_DIR)/$(APP_NAME)"

build-wasm:
	@echo "Building WebAssembly binary $(VERSION)..."
	@mkdir -p $(DIST_DIR)
	GOOS=js GOARCH=wasm $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME).wasm main.go
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

dist: build-wasm
	@echo "Preparing distribution files..."
	@mkdir -p $(DIST_DIR)
	@# Copy wasm_exec.js for reference, but we will also inline it
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(DIST_DIR)/
	@# Generate index.html with WebLLM support and inlined wasm_exec.js
	@echo "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>Oinakos</title><style>body { margin: 0; background: #000; overflow: hidden; display: flex; justify-content: center; align-items: center; height: 100vh; font-family: sans-serif; color: #daa520; flex-direction: column; } #status { font-size: 24px; margin-bottom: 10px; } #llm-status { font-size: 14px; color: #888; }</style></head><body><div id=\"status\">Loading Oinakos...</div><div id=\"llm-status\">Initializing Local WebGPU LLM...</div><script type=\"module\">" > $(DIST_DIR)/index.html
	@echo "import * as webllm from 'https://esm.run/@mlc-ai/web-llm';" >> $(DIST_DIR)/index.html
	@echo "window.oinakosWebLLM = {" >> $(DIST_DIR)/index.html
	@echo "  engine: null," >> $(DIST_DIR)/index.html
	@echo "  async init() {" >> $(DIST_DIR)/index.html
	@echo "    try {" >> $(DIST_DIR)/index.html
	@echo "      this.engine = await webllm.CreateMLCEngine('Llama-3-8B-Instruct-v0.1-q4f32_1-MLC', { initProgressCallback: (p) => { document.getElementById('llm-status').innerText = 'LLM: ' + p.text; } });" >> $(DIST_DIR)/index.html
	@echo "      document.getElementById('llm-status').innerText = 'Local LLM Ready (WebGPU)'; " >> $(DIST_DIR)/index.html
	@echo "    } catch (e) { console.error('WebGPU LLM failed:', e); document.getElementById('llm-status').innerText = 'Local LLM Failed (Using Noop)'; }" >> $(DIST_DIR)/index.html
	@echo "  }," >> $(DIST_DIR)/index.html
	@echo "  async chat(system, user, history) {" >> $(DIST_DIR)/index.html
	@echo "    if (!this.engine) return '...';" >> $(DIST_DIR)/index.html
	@echo "    const messages = [{role: 'system', content: system}, ...history, {role: 'user', content: user}];" >> $(DIST_DIR)/index.html
	@echo "    const reply = await this.engine.chat.completions.create({ messages });" >> $(DIST_DIR)/index.html
	@echo "    return reply.choices[0].message.content;" >> $(DIST_DIR)/index.html
	@echo "  }," >> $(DIST_DIR)/index.html
	@echo "  async decide(situation, options) {" >> $(DIST_DIR)/index.html
	@echo "    if (!this.engine) return {choice: options[0], reasoning: 'LLM not loaded'};" >> $(DIST_DIR)/index.html
	@echo "    const prompt = 'Situation: ' + situation + '\\\\nOptions: ' + options.join(', ') + '\\\\nPick one option and provide a short reasoning. Format: CHOICE: <option>\\\\nREASONING: <reasoning>';" >> $(DIST_DIR)/index.html
	@echo "    const reply = await this.engine.chat.completions.create({ messages: [{role: 'user', content: prompt}] });" >> $(DIST_DIR)/index.html
	@echo "    const text = reply.choices[0].message.content;" >> $(DIST_DIR)/index.html
	@echo "    const choiceMatch = text.match(/CHOICE:\\\\s*(.*)/i);" >> $(DIST_DIR)/index.html
	@echo "    const reasoningMatch = text.match(/REASONING:\\\\s*(.*)/i);" >> $(DIST_DIR)/index.html
	@echo "    return { choice: choiceMatch ? choiceMatch[1].trim() : options[0], reasoning: reasoningMatch ? reasoningMatch[1].trim() : '' };" >> $(DIST_DIR)/index.html
	@echo "  }" >> $(DIST_DIR)/index.html
	@echo "};" >> $(DIST_DIR)/index.html
	@echo "window.oinakosWebLLM.init();" >> $(DIST_DIR)/index.html
	@echo "</script><script>" >> $(DIST_DIR)/index.html
	@cat $(DIST_DIR)/wasm_exec.js >> $(DIST_DIR)/index.html
	@echo "</script><script>console.time('WASM Load'); const go = new Go(); WebAssembly.instantiateStreaming(fetch('oinakos.wasm'), go.importObject).then((result) => { console.timeEnd('WASM Load'); console.time('WASM Run'); document.getElementById('status').style.display = 'none'; go.run(result.instance); });</script></body></html>" >> $(DIST_DIR)/index.html
	rm $(DIST_DIR)/wasm_exec.js
	@echo "Dist files prepared in $(DIST_DIR)/: index.html (inlined JS) and oinakos.wasm"

run: build
	./$(BIN_DIR)/$(APP_NAME)

run-debug: build
	./$(BIN_DIR)/$(APP_NAME) -debug

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

clean:
	@echo "Cleaning up..."
	$(GOCLEAN)
	rm -rf $(BIN_DIR)
	rm -rf $(DIST_DIR)
	rm -rf tools/bin
	@echo "Cleaned."
