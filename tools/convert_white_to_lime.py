import os
from PIL import Image

def convert_to_lime(img_path):
    try:
        with Image.open(img_path) as img:
            img = img.convert("RGBA")
            data = img.getdata()
            new_data = []
            
            # Simple threshold for near-white
            for item in data:
                # item is (R, G, B, A)
                if item[3] > 200 and item[0] > 240 and item[1] > 240 and item[2] > 240:
                    # Replace with lime green (#00FF00) solid
                    new_data.append((0, 255, 0, 255))
                else:
                    new_data.append(item)
                    
            img.putdata(new_data)
            img.save(img_path, "PNG")
            print(f"Converted: {img_path}")
            return True
    except Exception as e:
        print(f"Error converting {img_path}: {e}")
        return False

files_to_convert = [
    "assets/images/archetypes/magi/male/corpse.png",
    "assets/images/archetypes/orc/male/static.png",
    "assets/images/archetypes/orc/male/corpse.png",
    "assets/images/archetypes/demon/male/static.png",
    "assets/images/archetypes/demon/male/corpse.png",
    "assets/images/environment/rock1.png",
    "assets/images/environment/bush5.png",
    "assets/images/environment/tree1.png",
    "assets/images/environment/bush2.png",
    "assets/images/environment/grass_tile.png",
    "assets/images/characters/main/static.png",
    "assets/images/characters/main/corpse.png",
    "assets/images/characters/main/attack.png"
]

for f in files_to_convert:
    convert_to_lime(f)

