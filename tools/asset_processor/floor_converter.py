import os
from PIL import Image, ImageDraw

def mask_to_diamond(img):
    """
    Creates a new 64x64 RGBA image with #00FF00 background.
    Resizes the given image and crops it to a 64x32 diamond centered horizontally.
    """
    # Create the 64x64 base with Lime Green
    out = Image.new("RGBA", (64, 64), (0, 255, 0, 255))
    
    # Resize the source image to cover the diamond area (or maybe just 64x64)
    # The diamond is from x=0 to 64, y=16 to 48.
    resized = img.resize((64, 64), Image.Resampling.LANCZOS).convert("RGBA")
    
    # Create a mask for the diamond
    mask = Image.new("L", (64, 64), 0)
    draw = ImageDraw.Draw(mask)
    # The diamond coordinates: (left, center_y), (center_x, top_y), (right, center_y), (center_x, bottom_y)
    draw.polygon([(0, 32), (32, 16), (64, 32), (32, 48)], fill=255)
    
    # Paste using the mask
    out.paste(resized, (0, 0), mask)
    
    return out

def process_floors(directory):
    for filename in os.listdir(directory):
        if filename.endswith(".png"):
            path = os.path.join(directory, filename)
            img = Image.open(path)
            out_img = mask_to_diamond(img)
            out_img.save(path, format="PNG")
            print(f"Processed {filename}")
            
if __name__ == "__main__":
    process_floors("../../assets/images/floors")
