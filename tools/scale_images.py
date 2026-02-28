import os
from PIL import Image

def scale_image(file_path, factor):
    if not os.path.exists(file_path):
        print(f"File not found: {file_path}")
        return
    
    with Image.open(file_path) as img:
        new_size = (int(img.width * factor), int(img.height * factor))
        print(f"Scaling {file_path}: {img.width}x{img.height} -> {new_size[0]}x{new_size[1]}")
        scaled_img = img.resize(new_size, Image.Resampling.LANCZOS)
        scaled_img.save(file_path)

def main():
    # Buildings to scale by 5.0
    buildings = [
        "assets/images/environment/house1.png",
        "assets/images/environment/house2.png",
        "assets/images/environment/house3.png",
        "assets/images/environment/smithery.png",
        "assets/images/environment/temple.png",
        "assets/images/environment/warehouse.png",
        "assets/images/environment/farm.png",
        "assets/images/environment/well.png"
    ]
    
    print("Scaling buildings...")
    for b in buildings:
        scale_image(b, 5.0)

if __name__ == "__main__":
    main()
