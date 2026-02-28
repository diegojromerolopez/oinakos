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

def get_character_images():
    output = subprocess.check_output(['find', 'assets/images/archetypes', 'assets/images/characters', '-name', '*.png'], text=True)
    return [line.strip() for line in output.splitlines() if line.strip()]

def main():
    # Scale NPCs and Main Character by 0.5
    print("Scaling NPCs and Characters by 0.5...")
    char_images = get_character_images()
    for img in char_images:
        scale_image(img, 0.5)

if __name__ == "__main__":
    main()
