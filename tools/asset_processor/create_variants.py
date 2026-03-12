import os
from PIL import Image, ImageEnhance

base_dir = "assets/images/floors"
target_files = [
    "grass.png",
    "mud.png",
    "paved_ground.png",
    "dirt.png",
    "big_stones.png",
    "wheat_field.png",
    "desert_sand.png",
    "yellow_grass.png",
    "dry_ground.png",
    "water.png",
    "dark_water.png"
]

for filename in target_files:
    filepath = os.path.join(base_dir, filename)
    if not os.path.exists(filepath):
        print(f"Skipping {filename}, not found.")
        continue
        
    img = Image.open(filepath)
    name, ext = os.path.splitext(filename)
    
    # Variant 2: Slightly darker / different color intensity
    enh_c = ImageEnhance.Color(img)
    img2 = enh_c.enhance(0.85)
    enh_b = ImageEnhance.Brightness(img2)
    img2 = enh_b.enhance(0.9)
    img2.save(os.path.join(base_dir, f"{name}_2{ext}"))
    
    # Variant 3: Slightly lighter
    enh_b2 = ImageEnhance.Brightness(img)
    img3 = enh_b2.enhance(1.1)
    img3.save(os.path.join(base_dir, f"{name}_3{ext}"))

print("Created variants for all floor types successfully!")
