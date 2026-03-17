import os
import glob
import yaml
from PIL import Image
import math

# Project constants (must match internal/engine/iso.go)
TILE_WIDTH = 64
TILE_HEIGHT = 32
HALF_W = TILE_WIDTH / 2
HALF_H = TILE_HEIGHT / 2

# Sprite Pivot (must match internal/game/item_instance.go)
SPRITE_W = 640
SPRITE_H = 640
PIVOT_X = SPRITE_W / 2
PIVOT_Y = SPRITE_H * 0.85

LIME_GREEN = (0, 255, 0)

def extract_perimeter(img_path):
    img = Image.open(img_path).convert("RGBA")
    pixels = img.load()
    width, height = img.size
    
    # Find all non-green pixels
    points = []
    for y in range(height):
        for x in range(width):
            r, g, b, a = pixels[x, y]
            if (r, g, b) != LIME_GREEN and a > 128:
                # Convert pixel to relative Cartesian
                screenX = x - PIVOT_X
                screenY = y - PIVOT_Y
                
                # Cartesian relative coordinates
                relX = (screenX / HALF_W + screenY / HALF_H) / 2
                relY = (screenY / HALF_H - screenX / HALF_W) / 2
                
                points.append((relX, relY))
    
    if not points:
        return []

    # Simplify points to a bounding box or simple hull
    # For now, let's just take a simplified bounding box in Cartesian space
    min_x = min(p[0] for p in points)
    max_x = max(p[0] for p in points)
    min_y = min(p[1] for p in points)
    max_y = max(p[1] for p in points)
    
    # Return 4 points of the bounding box
    return [
        {"x": round(min_x, 3), "y": round(min_y, 3)},
        {"x": round(max_x, 3), "y": round(min_y, 3)},
        {"x": round(max_x, 3), "y": round(max_y, 3)},
        {"x": round(min_x, 3), "y": round(max_y, 3)},
    ]

def process_all_objects():
    object_yamls = glob.glob("data/objects/*.yaml")
    for yaml_path in object_yamls:
        obj_id = os.path.splitext(os.path.basename(yaml_path))[0]
        img_path = os.path.join("assets/images/objects", f"{obj_id}.png")
        
        if not os.path.exists(img_path):
            print(f"Skipping {obj_id}: image not found")
            continue
            
        print(f"Processing {obj_id}...")
        footprint = extract_perimeter(img_path)
        
        if not footprint:
            print(f"Warning: No pixels found for {obj_id}")
            continue
            
        # Load existing YAML
        with open(yaml_path, "r") as f:
            data = yaml.safe_load(f)
            
        # Update footprint
        data["footprint"] = footprint
        
        # Save back
        with open(yaml_path, "w") as f:
            yaml.dump(data, f, sort_keys=False)

if __name__ == "__main__":
    process_all_objects()
