#!/bin/bash
set -e

APP_NAME="Oinakos"
BUNDLE_DIR="dist/${APP_NAME}.app"
CONTENTS_DIR="${BUNDLE_DIR}/Contents"
MAC_DIR="${CONTENTS_DIR}/MacOS"
RES_DIR="${CONTENTS_DIR}/Resources"

echo "Creating bundle structure..."
mkdir -p "${MAC_DIR}"
mkdir -p "${RES_DIR}"

echo "Building native binary for macOS..."
# Use -ldflags="-s -w" to reduce size and -X main.Version to inject version.
go build -ldflags "-s -w -X main.Version=${VERSION:-1.0}" -o "${MAC_DIR}/oinakos" main.go

echo "Generating Info.plist..."
cat > "${CONTENTS_DIR}/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleExecutable</key>
	<string>oinakos</string>
	<key>CFBundleIconFile</key>
	<string>icon.icns</string>
	<key>CFBundleIdentifier</key>
	<string>com.oinakos.game</string>
	<key>CFBundleName</key>
	<string>Oinakos</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleShortVersionString</key>
	<string>${VERSION:-1.0}</string>
	<key>CFBundleVersion</key>
	<string>1</string>
	<key>LSMinimumSystemVersion</key>
	<string>10.12</string>
	<key>NSHighResolutionCapable</key>
	<true/>
</dict>
</plist>
EOF

echo "Generating icon.icns..."
ICON_SRC="assets/images/characters/oinakos/static.png"
TEMP_ICON_DIR="/tmp/oinakos_icon.iconset"
mkdir -p "${TEMP_ICON_DIR}"

# Create transparent PNG first using our python script via uv (per GEMINI.md rules)
uv run tools/transparent_icon.py "${ICON_SRC}" "/tmp/transparent_base.png"

# Generate various sizes for iconset
# iconutil requires specifically named files
sips -z 16 16     /tmp/transparent_base.png --out "${TEMP_ICON_DIR}/icon_16x16.png" > /dev/null
sips -z 32 32     /tmp/transparent_base.png --out "${TEMP_ICON_DIR}/icon_16x16@2x.png" > /dev/null
sips -z 32 32    /tmp/transparent_base.png --out "${TEMP_ICON_DIR}/icon_32x32.png" > /dev/null
sips -z 64 64     /tmp/transparent_base.png --out "${TEMP_ICON_DIR}/icon_32x32@2x.png" > /dev/null
sips -z 128 128   /tmp/transparent_base.png --out "${TEMP_ICON_DIR}/icon_128x128.png" > /dev/null
sips -z 256 256   /tmp/transparent_base.png --out "${TEMP_ICON_DIR}/icon_128x128@2x.png" > /dev/null
sips -z 256 256   /tmp/transparent_base.png --out "${TEMP_ICON_DIR}/icon_256x256.png" > /dev/null
sips -z 512 512   /tmp/transparent_base.png --out "${TEMP_ICON_DIR}/icon_256x256@2x.png" > /dev/null
sips -z 512 512   /tmp/transparent_base.png --out "${TEMP_ICON_DIR}/icon_512x512.png" > /dev/null
sips -z 1024 1024 /tmp/transparent_base.png --out "${TEMP_ICON_DIR}/icon_512x512@2x.png" > /dev/null

iconutil -c icns "${TEMP_ICON_DIR}" -o "${RES_DIR}/icon.icns"

# Clean up
rm -rf "${TEMP_ICON_DIR}"
rm "/tmp/transparent_base.png"

echo "Bundle created: ${BUNDLE_DIR}"
echo "You can now run it using: open ${BUNDLE_DIR}"
