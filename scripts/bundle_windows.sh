#!/bin/bash
set -e

APP_NAME="oinakos"
RELEASE_NAME="Oinakos_Windows"
DIST_DIR="bin/${RELEASE_NAME}"
BIN_DIR="bin"

echo "Creating Windows distribution directory..."
mkdir -p "${DIST_DIR}"

echo "Building Windows binary..."
# -H=windowsgui hides the console window when the app runs
# CGO_ENABLED=0 ensures a portable binary without C dependencies
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-H=windowsgui" -o "${DIST_DIR}/${APP_NAME}.exe" main.go

echo "Generating Windows icon (ICO)..."
ICON_SRC="assets/images/characters/main/static.png"
uv run tools/transparent_icon.py "${ICON_SRC}" "${DIST_DIR}/icon.ico"

# Create a zip package
echo "Packaging into ZIP..."
cd "${BIN_DIR}"
zip -r "${RELEASE_NAME}.zip" "${RELEASE_NAME}"
cd ..

echo "Windows distribution created: bin/${RELEASE_NAME}.zip"
