import sys
from PIL import Image

def process(path):
    img = Image.open(path).convert("RGBA")
    data = img.getdata()
    new_data = []
    for item in data:
        if item[3] == 0: # fully transparent
            new_data.append((0, 255, 0, 255))
        else:
            new_data.append(item)
    img.putdata(new_data)
    img.save(path)
    print(f"Processed: {path}")

for path in sys.argv[1:]:
    process(path)
