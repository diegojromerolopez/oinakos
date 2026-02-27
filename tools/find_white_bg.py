import os
from PIL import Image

def has_white_background(img_path):
    try:
        with Image.open(img_path) as img:
            img = img.convert("RGBA")
            data = img.getdata()
            
            # Check the four corners for white or near-white
            w, h = img.size
            corners = [
                img.getpixel((0, 0)),
                img.getpixel((w-1, 0)),
                img.getpixel((0, h-1)),
                img.getpixel((w-1, h-1))
            ]
            
            # If any corner is near-white and fully opaque
            for r, g, b, a in corners:
                if a > 200 and r > 240 and g > 240 and b > 240:
                    return True
            return False
    except Exception as e:
        print(f"Error reading {img_path}: {e}")
        return False

def main():
    assets_dir = 'assets/images'
    white_bg_files = []
    
    for root, dirs, files in os.walk(assets_dir):
        for file in files:
            if file.endswith('.png'):
                full_path = os.path.join(root, file)
                if has_white_background(full_path):
                    white_bg_files.append(full_path)
    
    print(f"Found {len(white_bg_files)} images with white backgrounds:")
    for f in white_bg_files:
        print(f)

if __name__ == '__main__':
    main()
