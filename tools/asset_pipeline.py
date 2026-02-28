import os
import argparse
import sys
from PIL import Image

def process_image(input_path, output_path, config):
    if not os.path.exists(input_path):
        print(f"Error: Could not find image at {input_path}")
        return False
        
    try:
        img = Image.open(input_path).convert('RGB')
        
        # Determine if we should crop by scanning for the background color
        if config.get('auto_crop', False):
            bg_color = config.get('bg_color', (0, 255, 0)) # default lime
            
            width, height = img.size
            min_x, min_y, max_x, max_y = width, height, 0, 0
            
            pixels = img.load()
            for y in range(height):
                for x in range(width):
                    r, g, b = pixels[x, y]
                    
                    # Logic: is the pixel NOT the background color?
                    is_bg = abs(r - bg_color[0]) < 20 and abs(g - bg_color[1]) < 20 and abs(b - bg_color[2]) < 20
                    if not is_bg:
                        if x < min_x: min_x = x
                        if x > max_x: max_x = x
                        if y < min_y: min_y = y
                        if y > max_y: max_y = y
            
            # Valid object found if bounds make sense
            if min_x <= max_x and min_y <= max_y:
                img = img.crop((min_x, min_y, max_x, max_y))
        
        # Apply scaling if an exact size or scale factor was provided
        if config.get('target_size'):
            img = img.resize(config['target_size'], Image.Resampling.LANCZOS)
        elif config.get('scale_factor'):
            sf = config['scale_factor']
            w, h = img.size
            img = img.resize((max(1, int(w * sf)), max(1, int(h * sf))), Image.Resampling.LANCZOS)
            
        # Optional: Pad to strictly match a canvas size containing a specific background
        if config.get('canvas_size') and config.get('bg_color'):
            cw, ch = config['canvas_size']
            bg_color = config['bg_color']
            
            final_img = Image.new('RGB', (cw, ch), bg_color)
            iw, ih = img.size
            
            # Center the image
            offset_x = max(0, (cw - iw) // 2)
            offset_y = max(0, (ch - ih) // 2)
            final_img.paste(img, (offset_x, offset_y))
            img = final_img
            
        # Optional: Flood-fill background completely based on another mask file to fix artifacts
        if config.get('mask_file') and config.get('bg_color'):
            mask_path = config['mask_file']
            if os.path.exists(mask_path):
                mask_img = Image.open(mask_path).convert('RGB')
                mask_pixels = mask_img.load()
                
                # Assume original mask BG is what we want to project
                img_pixels = img.load()
                w, h = min(img.size[0], mask_img.size[0]), min(img.size[1], mask_img.size[1])
                
                for y in range(h):
                    for x in range(w):
                        mr, mg, mb = mask_pixels[x, y]
                        # Assume mask background is lime
                        if mr < 20 and mg > 230 and mb < 20: 
                            img_pixels[x, y] = config['bg_color']
            
        # Save output
        out_dir = os.path.dirname(output_path)
        if out_dir and not os.path.exists(out_dir):
            os.makedirs(out_dir)
            
        img.save(output_path)
        print(f"Successfully processed {input_path} -> {output_path}")
        return True
        
    except Exception as e:
        print(f"Error processing {input_path}: {e}")
        return False


def main():
    parser = argparse.ArgumentParser(description="A generic tool pipeline for automated asset resizing, cropping, and masking.")
    parser.add_argument('input', help='Path to the input image file')
    parser.add_argument('--output', '-o', help='Path to save the output image (overwrites input if omitted)')
    
    # Operations
    parser.add_argument('--size', '-s', type=int, nargs=2, metavar=('WIDTH', 'HEIGHT'), help='Absolute target resize dimensions (w h)')
    parser.add_argument('--scale', type=float, help='A multiplier scale factor to resize by')
    parser.add_argument('--canvas', type=int, nargs=2, metavar=('WIDTH', 'HEIGHT'), help='Creates a specific canvas bounding-box size')
    parser.add_argument('--auto-crop', action='store_true', help='Strips out the target background color around the object')
    parser.add_argument('--mask', help='Path to an image whose background to copy layout from (fixes artifacts)')
    
    # Colors
    parser.add_argument('--bg-color', type=int, nargs=3, default=[0, 255, 0], metavar=('R', 'G', 'B'), help='RGB background matching color (default: 0 255 0 lime green)')

    args = parser.parse_args()
    
    config = {
        'bg_color': tuple(args.bg_color),
        'auto_crop': args.auto_crop,
    }
    
    if args.size:
        config['target_size'] = tuple(args.size)
    if args.scale:
        config['scale_factor'] = args.scale
    if args.canvas:
        config['canvas_size'] = tuple(args.canvas)
    if args.mask:
        config['mask_file'] = args.mask
        
    dest = args.output if args.output else args.input
    process_image(args.input, dest, config)

if __name__ == "__main__":
    main()
