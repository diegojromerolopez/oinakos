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
            
            # Project standard: Solid lime green background (#00FF00)
            # Replace white OR transparent areas with lime
            for item in data:
                # item is (R, G, B, A)
                # If transparent (A < 50) OR white (R,G,B > 240)
                if item[3] < 50 or (item[0] > 240 and item[1] > 240 and item[2] > 240):
                    new_data.append((0, 255, 0, 255))
                else:
                    new_data.append(item)
                    
            img.putdata(new_data)
            img.save(img_path, "PNG")
            print(f"Converted to Lime: {img_path}")
            return True
    except Exception as e:
        print(f"Error converting {img_path}: {e}")
        return False

if __name__ == "__main__":
    targets = [
        "assets/images/environment/tree2.png",
        "assets/images/environment/bush1.png",
        "assets/images/environment/bush2.png"
    ]
    for t in targets:
        convert_to_lime(t)
