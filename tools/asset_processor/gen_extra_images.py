import sys
import os
from PIL import Image, ImageEnhance, ImageChops

def extract_char(image):
    bg_color = (0, 255, 0)
    data = []
    # Avoid warning for old PILLOW versions manually or just use getdata
    pixel_data = list(image.getdata()) if hasattr(image, 'getdata') else list(image.get_flattened_data())
    for pixel in pixel_data:
        if pixel[:3] == bg_color:
            data.append((0, 255, 0, 0)) # transparent internally
        else:
            data.append(pixel)
    tmp = Image.new("RGBA", image.size)
    tmp.putdata(data)
    return tmp

def finalize_img(image):
    bg = Image.new("RGBA", image.size, (0, 255, 0, 255))
    bg.paste(image, (0, 0), image)
    return bg

def create_hit(c_img):
    hit_data = []
    pixel_data = list(c_img.getdata()) if hasattr(c_img, 'getdata') else list(c_img.get_flattened_data())
    for r, g, b, a in pixel_data:
        if a == 0:
            hit_data.append((0, 255, 0, 0))
        else:
            hit_data.append((min(255, int(r*1.5)), min(255, int(g*0.7)), min(255, int(b*0.7)), a))
    tmp = Image.new("RGBA", c_img.size)
    tmp.putdata(hit_data)
    shifted = Image.new("RGBA", c_img.size, (0, 255, 0, 0))
    shifted.paste(tmp, (5, -5), tmp)
    return finalize_img(shifted)

def create_corpse(c_img):
    rot = c_img.rotate(90, expand=False)
    enhancer = ImageEnhance.Brightness(rot)
    rot = enhancer.enhance(0.4)
    moved = Image.new("RGBA", c_img.size, (0, 255, 0, 0))
    # Move downward based on image size to prevent cutting off
    moved.paste(rot, (0, 40), rot)
    return finalize_img(moved)

def create_attack(c_img, dx, dy):
    moved = Image.new("RGBA", c_img.size, (0, 255, 0, 0))
    moved.paste(c_img, (dx, dy), c_img)
    return finalize_img(moved)

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python gen_extra_images.py path/to/static.png")
        sys.exit(1)
        
    static_path = sys.argv[1]
    base_dir = os.path.dirname(static_path)
    if not os.path.exists(static_path):
        print(f"Error: {static_path} not found.")
        sys.exit(1)

    img = Image.open(static_path).convert("RGBA")
    char_img = extract_char(img)

    create_hit(char_img).save(os.path.join(base_dir, "hit.png"))
    create_corpse(char_img).save(os.path.join(base_dir, "corpse.png"))
    create_attack(char_img, 10, -10).save(os.path.join(base_dir, "attack.png"))
    
    # Check if we should also save attack1/attack2
    create_attack(char_img, 10, -10).save(os.path.join(base_dir, "attack1.png"))
    create_attack(char_img, -10, -10).save(os.path.join(base_dir, "attack2.png"))

    print(f"Generated extra images for {base_dir}")
