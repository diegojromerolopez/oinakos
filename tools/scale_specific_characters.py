import os
import subprocess
from PIL import Image

def scale_image(file_path, factor):
    if not os.path.exists(file_path):
        print(f"File not found: {file_path}")
        return
    
    with Image.open(file_path) as img:
        new_size = (int(img.width * factor), int(img.height * factor))
        if new_size[0] < 1 or new_size[1] < 1:
            print(f"Skipping {file_path}: too small")
            return
        print(f"Scaling {file_path}: {img.width}x{img.height} -> {new_size[0]}x{new_size[1]}")
        scaled_img = img.resize(new_size, Image.Resampling.LANCZOS)
        scaled_img.save(file_path)

def get_images_in_dir(path):
    if not os.path.exists(path):
        return []
    images = []
    for root, dirs, files in os.walk(path):
        for file in files:
            if file.endswith(".png"):
                images.append(os.path.join(root, file))
    return images

def main():
    # Scale Goblins and Magi by 0.5
    print("Scaling Goblins and Magi by 0.5...")
    goblins_m = get_images_in_dir("assets/images/archetypes/goblin")
    magi_m = get_images_in_dir("assets/images/archetypes/magi")
    for img in goblins_m + magi_m:
        scale_image(img, 0.5)

    # Scale Giants by 2.0
    print("\nScaling Giants by 2.0...")
    giants_m = get_images_in_dir("assets/images/archetypes/giant")
    for img in giants_m:
        scale_image(img, 2.0)

if __name__ == "__main__":
    main()
