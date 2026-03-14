#!/bin/bash
set -e

APP_NAME="oinakos"
RELEASE_NAME="Oinakos_Linux"
DIST_DIR="dist/${RELEASE_NAME}"
BIN_DIR="dist"

echo "Creating Linux distribution directory..."
mkdir -p "${DIST_DIR}"

echo "Building Linux binary via custom Docker image..."
echo "(This ensures native X11/OpenGL/ALSA headers are present for Ebiten)"

# Build our specialized build image
docker build -t oinakos-builder -f scripts/Dockerfile.linux .

# Run the build
docker run --rm \
    -v "$(pwd)":/src \
    -w /src \
    -e CGO_ENABLED=1 \
    -e VERSION="${VERSION:-0.1}" \
    oinakos-builder \
    go build -ldflags "-X main.Version=${VERSION:-0.1}" -o "dist/Oinakos_Linux/oinakos" main.go

if [ ! -f "${DIST_DIR}/${APP_NAME}" ]; then
    echo "ERROR: Linux build failed within Docker container."
    exit 1
fi

echo "Generating Linux icon (PNG)..."
ICON_SRC="assets/images/characters/oinakos/static.png"
uv run tools/transparent_icon.py "${ICON_SRC}" "${DIST_DIR}/icon.png"

echo "Generating Desktop Entry..."
cat > "${DIST_DIR}/oinakos.desktop" <<EOF
[Desktop Entry]
Name=Oinakos
Comment=Isometric Action RPG
Exec=./oinakos
Icon=./icon.png
Terminal=false
Type=Application
Categories=Game;
EOF

# Create a tarball
echo "Packaging into Tarball..."
cd "${BIN_DIR}"
tar -czf "${RELEASE_NAME}.tar.gz" "${RELEASE_NAME}"
cd ..

echo "Linux distribution created: dist/${RELEASE_NAME}.tar.gz"
