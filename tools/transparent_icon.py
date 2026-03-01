import sys
from PIL import Image

def transparentize(img_path, out_path):
    img = Image.open(img_path).convert("RGBA")
    datas = img.getdata()

    newData = []
    for item in datas:
        r, g, b, a = item
        # Robust lime detection
        if g > 140 and g > (r * 1.2) and g > (b * 2):
            newData.append((0, 0, 0, 0))
        else:
            newData.append(item)

    img.putdata(newData)
    
    if out_path.endswith(".ico"):
        # Windows ICO supports multiple sizes in one file
        icon_sizes = [(16, 16), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)]
        img.save(out_path, format="ICO", sizes=icon_sizes)
    else:
        img.save(out_path, "PNG")

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python transparent_icon.py input.png output.png")
        sys.exit(1)
    transparentize(sys.argv[1], sys.argv[2])
