import os
from PIL import Image

def scale_image(file_path, factor):
    if not os.path.exists(file_path):
        print(f"File not found: {file_path}")
        return
    
    with Image.open(file_path) as img:
        # We want to scale relative to the ORIGINAL size (640x640)
        # Assuming they are currently at 640x640 (after my last restore)
        new_size = (int(img.width * factor), int(img.height * factor))
        print(f"Scaling {file_path}: {img.width}x{img.height} -> {new_size[0]}x{new_size[1]}")
        scaled_img = img.resize(new_size, Image.Resampling.LANCZOS)
        scaled_img.save(file_path)

def main():
    # Nature elements to scale down by 0.5
    nature = [
        "assets/images/environment/tree1.png",
        "assets/images/environment/tree2.png",
        "assets/images/environment/rock1.png",
        "assets/images/environment/rock2.png",
        "assets/images/environment/rock3.png",
        "assets/images/environment/rock4.png",
        "assets/images/environment/rock5.png",
        "assets/images/environment/bush1.png",
        "assets/images/environment/bush2.png",
        "assets/images/environment/bush3.png",
        "assets/images/environment/bush4.png",
        "assets/images/environment/bush5.png",
    ]

    print("Scaling nature elements to half size (0.5)...")
    for n in nature:
        scale_image(n, 0.5)

if __name__ == "__main__":
    main()
