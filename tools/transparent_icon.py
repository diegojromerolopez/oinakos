#!/usr/bin/env python3
"""
transparent_icon.py — Convert a chroma-key green (#00FF00) sprite to a
transparent PNG suitable for use as a macOS .icns icon source.

Usage:
    uv run tools/transparent_icon.py <input.png> <output.png>
"""

# /// script
# requires-python = ">=3.10"
# dependencies = ["Pillow"]
# ///

import sys
from pathlib import Path
from PIL import Image


def remove_chromakey(input_path: str, output_path: str) -> None:
    """Replace chroma-key green (#00FF00) pixels with transparency."""
    img = Image.open(input_path).convert("RGBA")
    pixels = img.load()

    width, height = img.size
    for y in range(height):
        for x in range(width):
            r, g, b, a = pixels[x, y]
            # Match pure chroma-key green and near-variants (tolerance ±15)
            if g > 200 and r < 80 and b < 80:
                pixels[x, y] = (0, 0, 0, 0)

    img.save(output_path, "PNG")
    print(f"Saved transparent icon: {output_path}")


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print(f"Usage: {sys.argv[0]} <input.png> <output.png>", file=sys.stderr)
        sys.exit(1)

    input_file = sys.argv[1]
    output_file = sys.argv[2]

    if not Path(input_file).exists():
        print(f"Error: input file not found: {input_file}", file=sys.stderr)
        sys.exit(1)

    remove_chromakey(input_file, output_file)
