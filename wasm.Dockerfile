FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for private repos if needed
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code explicitly
COPY assets/ ./assets/
COPY data/ ./data/
COPY internal/ ./internal/
COPY cmd/ ./cmd/
COPY main.go .
COPY go.mod go.sum ./

# Build the WASM binary
ENV GOOS=js
ENV GOARCH=wasm
RUN mkdir -p bin && go build -o bin/oinakos.wasm main.go

# Build the Go server binary (Native)
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
RUN go build -o bin/server cmd/server/main.go

# Copy wasm_exec.js from the Go distribution
RUN cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" bin/

# Serve stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the build artifacts from the builder stage
COPY --from=builder /app/bin/oinakos.wasm .
COPY --from=builder /app/assets/wasm/index.html .
COPY --from=builder /app/bin/wasm_exec.js .
COPY --from=builder /app/bin/server .

# Expose port 8000
EXPOSE 8000

# Run the Go server
CMD ["./server"]
