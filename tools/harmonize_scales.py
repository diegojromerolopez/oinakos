import os
from PIL import Image

def scale_image(file_path, target_width):
    if not os.path.exists(file_path):
        return
    try:
        with Image.open(file_path) as img:
            if img.width == target_width:
                return
            factor = target_width / img.width
            new_size = (int(img.width * factor), int(img.height * factor))
            if new_size[0] < 1 or new_size[1] < 1:
                return
            print(f"Resizing {file_path}: {img.width}x{img.height} -> {new_size[0]}x{new_size[1]}")
            img.resize(new_size, Image.Resampling.LANCZOS).save(file_path)
    except Exception as e:
        print(f"Error {file_path}: {e}")

def get_images(dir_path):
    imgs = []
    for root, dirs, files in os.walk(dir_path):
        for f in files:
            if f.endswith(".png"):
                imgs.append(os.path.join(root, f))
    return imgs

def main():
    # Humans/Normal NPCs -> 160px
    for d in ["assets/images/characters/main", "assets/images/archetypes/orc", 
              "assets/images/archetypes/demon", "assets/images/archetypes/peasant",
              "assets/images/archetypes/lame_devil", "assets/images/archetypes/goblin",
              "assets/images/archetypes/magi"]:
        for img in get_images(d):
            scale_image(img, 160)

    # Giants -> 800px
    for d in ["assets/images/archetypes/giant"]:
        for img in get_images(d):
            scale_image(img, 800)
            
    # Nature -> 320px
    nature_imgs = get_images("assets/images/environment")
    for img in nature_imgs:
        base = os.path.basename(img)
        if any(x in base for x in ["tree", "rock", "bush", "grass"]):
            scale_image(img, 320)
            
    # Houses -> 1200px (approx 7.5x Human)
    for img in nature_imgs:
        base = os.path.basename(img)
        if any(x in base for x in ["house", "farm", "smithery", "temple", "warehouse"]):
            scale_image(img, 1200)

    # Well -> 300px
    scale_image("assets/images/environment/well.png", 300)

if __name__ == "__main__":
    main()
