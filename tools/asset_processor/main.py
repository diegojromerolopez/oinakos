from PIL import Image
import sys
import os

def process_npc_asset(in_path, out_path, force_magenta_band=False):
    img = Image.open(in_path).convert("RGBA")
    
    # Scale to 160x160
    img = img.resize((160, 160), Image.Resampling.LANCZOS)
    
    data = img.getdata()
    new_data = []
    
    # Simple background replacement
    # We look for colors that are NOT dark (characters are usually dark in this style)
    # or that are close to the background color the generator used.
    
    # If the user says solid lime green, let's assume the generator might try but fail to be exact.
    # We'll use a flood-fill or just proximity.
    
    # Actually, for Hades/Diablo style, character is central.
    # Let's try to detect the "most likely background" color from the corners.
    bg_color = data[0] # Top left pixel
    
    for item in data:
        r, g, b, a = item
        
        # If color is close to corner color or very bright (like a white/grey background)
        # and not Magenta (if we are looking for a band) or Yellow.
        
        # distance to bg_color
        dist = ((r - bg_color[0])**2 + (g - bg_color[1])**2 + (b - bg_color[2])**2)**0.5
        
        is_bg = dist < 50 # Threshold
        
        # Also check for "near white" backgrounds
        if r > 240 and g > 240 and b > 240:
            is_bg = True
            
        if is_bg:
            new_data.append((0, 255, 0, 255)) # PURE LIME
        else:
            # If we need to force the magenta band
            if force_magenta_band:
                # Look for colors that might be the band (purple-ish)
                # Primary color Magenta is (255, 0, 255)
                # If color is closer to Magenta than to anything else in the character?
                # This is risky. 
                # Let's just keep the color if we can.
                new_data.append(item)
            else:
                new_data.append(item)
                
    img.putdata(new_data)
    
    # Final check: Ensure Magenta #FF00FF and Yellow #FFFF00 are exactly those values 
    # if they are close enough (to fix generator blurriness).
    # This is for the palette swap feature.
    
    final_data = []
    for item in new_data:
        r, g, b, a = item
        if a == 0:
            final_data.append(item)
            continue
            
        # If it's #00FF00 (background), keep it
        if r == 0 and g == 255 and b == 0:
            final_data.append(item)
            continue
            
        # Force Magenta
        if abs(r - 255) < 30 and abs(g - 0) < 30 and abs(b - 255) < 30:
            final_data.append((255, 0, 255, 255))
        # Force Yellow
        elif abs(r - 255) < 30 and abs(g - 255) < 30 and abs(b - 0) < 30:
            final_data.append((255, 255, 0, 255))
        else:
            final_data.append(item)
            
    img.putdata(final_data)
    img.save(out_path)

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python process_npc.py in.png out.png [force_magenta]")
        sys.exit(1)
    
    force_mag = len(sys.argv) > 3 and sys.argv[3].lower() == "true"
    process_npc_asset(sys.argv[1], sys.argv[2], force_mag)
