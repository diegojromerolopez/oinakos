from PIL import Image, ImageDraw
import sys

def restore_lime_bg(img_path):
    print(f"Processing {img_path}...")
    try:
        img = Image.open(img_path).convert('RGB')
        width, height = img.size
        
        # Start flood fill from the 4 corners of the image
        corners = [(0, 0), (width - 1, 0), (0, height - 1), (width - 1, height - 1)]
        lime = (0, 255, 0)
        
        for corner in corners:
            # thresh=15 allows for slight jpeg/compression artifacts 
            # while protecting the main tile block
            ImageDraw.floodfill(img, corner, lime, thresh=15)
            
        img.save(img_path)
        print(f"Successfully restored background for {img_path}")
    except Exception as e:
        print(f"Error processing {img_path}: {e}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python floodfill_corners.py <image1> <image2> ...")
        sys.exit(1)
        
    for path in sys.argv[1:]:
        restore_lime_bg(path)
