import sys
import yaml
from PIL import Image
import numpy as np

def extract_perimeter(img_path):
    # Open image and convert to RGBA
    img = Image.open(img_path).convert("RGBA")
    data = np.array(img)
    
    # Standard lime green: R=0, G=255, B=0
    # But let's use the same robust logic as Go:
    # g8 > 140 && g8 > (r8 * 1.2) && g8 > (b8 * 2)
    r, g, b, a = data[:,:,0], data[:,:,1], data[:,:,2], data[:,:,3]
    
    # Mask of pixels that are NOT lime and are NOT transparent
    is_lime = (g > 140) & (g > (r.astype(float) * 1.2)) & (g > (b.astype(float) * 2))
    is_visible = (a > 0) & (~is_lime)
    
    coords = np.argwhere(is_visible)
    if len(coords) == 0:
        return []

    # Get the convex hull or just the bounding box for now?
    # Actually, the user wants a "polygon that determines the perimeter".
    # Since it's for 2D collision in world-space, we want a simplified perimeter.
    
    # Let's find the outermost points at different angles
    # Center of mass
    cy, cx = np.mean(coords, axis=0)
    
    # We'll sample 16 points around the perimeter
    perimeter_points = []
    num_samples = 16
    for i in range(num_samples):
        angle = 2 * np.pi * i / num_samples
        dx, dy = np.cos(angle), np.sin(angle)
        
        # Ray cast from center
        max_dist = 0
        best_pt = None
        for y, x in coords:
            # Vector from center to point
            vx, vy = x - cx, y - cy
            # Project onto angle vector
            dist = vx * dx + vy * dy
            if dist > max_dist:
                max_dist = dist
                best_pt = (x, y)
        
        if best_pt:
            perimeter_points.append(best_pt)

    # Convert to offsets from center
    # Image size
    h, w = data.shape[:2]
    off_points = []
    for x, y in perimeter_points:
        # We want world-space relative to center of image
        off_points.append({
            'x': float(x - w/2),
            'y': float(y - h/2)
        })
        
    return off_points

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: uv run tools/extract_perimeter.py path/to/sprite.png")
        sys.exit(1)
        
    pts = extract_perimeter(sys.argv[1])
    print(yaml.dump(pts, default_flow_style=False))
