import os
from PIL import Image

def process_arrow(input_path, output_path, target_size=64):
    if not os.path.exists(input_path):
        print(f'Missing {input_path}')
        return
    img = Image.open(input_path).convert('RGB')
    
    # Simple green screening bounding box
    bg_color = (0, 255, 0)
    width, height = img.size
    min_x, min_y, max_x, max_y = width, height, 0, 0
    
    pixels = img.load()
    for y in range(height):
        for x in range(width):
            r, g, b = pixels[x, y]
            if not (r < 50 and g > 200 and b < 50): # Not bright lime green
                if x < min_x: min_x = x
                if x > max_x: max_x = x
                if y < min_y: min_y = y
                if y > max_y: max_y = y
                
    if min_x > max_x or min_y > max_y:
        print('No object found in arrow image.')
        img = img.resize((target_size, target_size), Image.Resampling.LANCZOS)
    else:
        cropped = img.crop((min_x, min_y, max_x, max_y))
        
        crop_w, crop_h = cropped.size
        scale = target_size / max(crop_w, crop_h)
        new_w = max(1, int(crop_w * scale))
        new_h = max(1, int(crop_h * scale))
        img = cropped.resize((new_w, new_h), Image.Resampling.LANCZOS)
        
        final_img = Image.new('RGB', (target_size, target_size), (0, 255, 0))
        offset_x = (target_size - new_w) // 2
        offset_y = (target_size - new_h) // 2
        final_img.paste(img, (offset_x, offset_y))
        img = final_img
        
    img.save(output_path)
    print(f'Processed arrow to {output_path}')

artifacts = [
    ('/Users/diegoj/.gemini/antigravity/brain/2f1cbbd7-0bda-4a88-bd87-3f24744d3c57/bush2_dark_1772259859632.png', 'assets/images/environment/bush2.png'),
    ('/Users/diegoj/.gemini/antigravity/brain/2f1cbbd7-0bda-4a88-bd87-3f24744d3c57/bush3_dark_1772259871450.png', 'assets/images/environment/bush3.png'),
    ('/Users/diegoj/.gemini/antigravity/brain/2f1cbbd7-0bda-4a88-bd87-3f24744d3c57/bush4_dark_1772259916576.png', 'assets/images/environment/bush4.png'),
    ('/Users/diegoj/.gemini/antigravity/brain/2f1cbbd7-0bda-4a88-bd87-3f24744d3c57/bush5_dark_1772259883994.png', 'assets/images/environment/bush5.png')
]

for src, dst in artifacts:
    if os.path.exists(src):
        img = Image.open(src).convert('RGB')
        img = img.resize((150, 150), Image.Resampling.LANCZOS)
        img.save(dst)
        print(f'Processed {dst}')

process_arrow('/Users/diegoj/.gemini/antigravity/brain/2f1cbbd7-0bda-4a88-bd87-3f24744d3c57/medieval_arrow_1772259769887.png', 'assets/images/environment/arrow.png', 48)
