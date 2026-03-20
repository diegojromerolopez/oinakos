import os
from PIL import Image
import yaml

# Standard YAML library for Python. We'll try to preserve some formatting.
# If you want to keep comments, ruamel.yaml is better, but this should work for data.

IMAGES_DIR = "../../assets/images/obstacles"
DATA_DIR = "../../data/obstacles"
MAX_WIDTH = 512  # More aggressive limit (8x tile width)

def process_obstacles():
    print(f"Scanning {IMAGES_DIR}...")
    
    for filename in os.listdir(IMAGES_DIR):
        if not filename.endswith(".png"):
            continue
            
        name = filename[:-4]
        image_path = os.path.join(IMAGES_DIR, filename)
        yaml_path = os.path.join(DATA_DIR, name + ".yaml")
        
        if not os.path.exists(yaml_path):
            # print(f"Skipping {name}: no YAML found")
            continue
            
        try:
            with Image.open(image_path) as img:
                orig_w, orig_h = img.size
                
                # We care about the width of a single frame
                # Need to check YAML for frame_count
                frame_count = 1
                try:
                    with open(yaml_path, 'r') as f:
                        data = yaml.safe_load(f)
                        frame_count = data.get('frame_count', 1)
                except Exception as e:
                    print(f"Error reading YAML for {name}: {e}")
                    continue
                
                frame_w = orig_w
                if frame_count > 1:
                    # simplistic: assume single row for now or check frames_per_row
                    fpr = data.get('frames_per_row', frame_count)
                    if fpr <= 0: fpr = frame_count
                    frame_w = orig_w // fpr

                if frame_w > MAX_WIDTH:
                    # Calculate new dimensions
                    ratio = MAX_WIDTH / frame_w
                    new_w = int(orig_w * ratio)
                    new_h = int(orig_h * ratio)
                    
                    print(f"Downscaling {name}: {orig_w}x{orig_h} -> {new_w}x{new_h} (Ratio: {1/ratio:.2f}x)")
                    
                    # Resize
                    new_img = img.resize((new_w, new_h), Image.Resampling.LANCZOS)
                    new_img.save(image_path, optimize=True)
                    
                    # Update YAML
                    new_scale = 1.0 / ratio
                    try:
                        # Re-read to ensure we have all data
                        with open(yaml_path, 'r') as f:
                            lines = f.readlines()
                        
                        # We want to add/update 'scale: value'
                        # To avoid messing up YAML formatting with 'yaml.dump', 
                        # we'll do a simple string replacement if possible or just append.
                        
                        found_scale = False
                        old_scale = 1.0
                        new_lines = []
                        for line in lines:
                            if line.strip().startswith('scale:'):
                                try:
                                    old_scale = float(line.split(':')[1].strip())
                                except:
                                    pass
                                found_scale = True
                                # We'll replace it later
                            else:
                                new_lines.append(line)
                        
                        total_scale = old_scale / ratio
                        
                        scale_line = f"scale: {total_scale:.3f}\n"
                        if found_scale:
                            # Re-insert at reasonable position
                            new_lines.insert(len(new_lines) if not any(l.startswith('footprint:') for l in new_lines) else [i for i,l in enumerate(new_lines) if l.startswith('footprint:')][0], scale_line)
                        else:
                            # Add before footprint if exists, or at the end
                            inserted = False
                            for i, line in enumerate(new_lines):
                                if line.strip().startswith('footprint:'):
                                    new_lines.insert(i, scale_line)
                                    inserted = True
                                    break
                            if not inserted:
                                new_lines.append(scale_line)
                                
                        with open(yaml_path, 'w') as f:
                            f.writelines(new_lines)
                            
                        print(f"  Updated {yaml_path} with scale: {new_scale:.3f}")
                        
                    except Exception as e:
                        print(f"Error updating YAML for {name}: {e}")
                else:
                    # Even if we don't resize, we can optimize the PNG to save space
                    print(f"Optimizing {name} (no resize needed)")
                    img.save(image_path, optimize=True)
                    
        except Exception as e:
            print(f"Error processing {name}: {e}")

if __name__ == "__main__":
    process_obstacles()
