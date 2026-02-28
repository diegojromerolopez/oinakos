from PIL import Image
import os

def convert_to_lime(img_path):
    try:
        if not os.path.exists(img_path):
            print(f"File not found: {img_path}")
            return False
            
        with Image.open(img_path) as img:
            img = img.convert("RGBA")
            data = img.getdata()
            new_data = []
            
            # Simple threshold for white or near-white
            for item in data:
                # item is (R, G, B, A)
                # If R, G, B are all high and it's opaque, it's probably background
                if item[3] > 0 and item[0] > 240 and item[1] > 240 and item[2] > 240:
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

if __name__ == "__main__":
    convert_to_lime("assets/images/environment/tree2.png")
