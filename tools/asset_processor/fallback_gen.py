import sys
import os
from PIL import Image, ImageEnhance, ImageChops

def extract_char(image):
    bg_color = (0, 255, 0)
    data = []
    pixel_data = list(image.getdata()) if hasattr(image, 'getdata') else list(image.get_flattened_data())
    for pixel in pixel_data:
        if pixel[:3] == bg_color:
            data.append((0, 255, 0, 0))
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
    rot = c_img.rotate(85, expand=False)
    enhancer = ImageEnhance.Brightness(rot)
    rot = enhancer.enhance(0.4)
    moved = Image.new("RGBA", c_img.size, (0, 255, 0, 0))
    moved.paste(rot, (0, 45), rot)
    return finalize_img(moved)

def create_attack(c_img, dx, dy):
    moved = Image.new("RGBA", c_img.size, (0, 255, 0, 0))
    moved.paste(c_img, (dx, dy), c_img)
    return finalize_img(moved)

def create_static(c_img):
    return finalize_img(c_img)

if __name__ == "__main__":
    v_base_path = "assets/images/npcs/virculus/static.png"
    if os.path.exists(v_base_path):
        img_v = Image.open(v_base_path).convert("RGBA")
        base_v = extract_char(img_v)
        # Create attack2 and attack3
        # Virculus attack2: Lunge deeper
        create_attack(base_v, 15, -5).save("assets/images/npcs/virculus/attack2.png")
        # Virculus attack3: Large downward shift (smash)
        create_attack(base_v, 5, 20).save("assets/images/npcs/virculus/attack3.png")



    print("Procedural generation complete.")
