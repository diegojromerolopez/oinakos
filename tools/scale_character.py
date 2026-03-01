import os
from PIL import Image

def scale_image(input_path, output_path, scale_factor=0.75):
    img = Image.open(input_path).convert("RGBA")
    w, h = img.size
    
    # Calculate new size
    new_w = int(w * scale_factor)
    new_h = int(h * scale_factor)
    
    # Scale character
    scaled_img = img.resize((new_w, new_h), Image.Resampling.LANCZOS)
    
    # Create new transparent canvas
    new_canvas = Image.new("RGBA", (w, h), (0, 255, 0, 255))
    
    # Paste scaled character centered
    paste_x = (w - new_w) // 2
    paste_y = (h - new_h) // 2
    
    # If the original image had transparency represented by lime green, 
    # we should handle it. But the prompt said the background is lime green.
    # We'll just paste the scaled image on the new lime green canvas.
    new_canvas.paste(scaled_img, (paste_x, paste_y), scaled_img)
    
    new_canvas.save(output_path)
    print(f"Scaled {input_path} to {output_path}")

if __name__ == "__main__":
    # Source paths (the original high-res generated images)
    sources = {
        "static.png": "/Users/diegoj/.gemini/antigravity/brain/5203640c-9129-4122-886c-4ff7a903e705/knight_static_hades_style_1772400491688.png",
        "attack.png": "/Users/diegoj/.gemini/antigravity/brain/5203640c-9129-4122-886c-4ff7a903e705/knight_attack_hades_style_1772400503281.png",
        "corpse.png": "/Users/diegoj/.gemini/antigravity/brain/5203640c-9129-4122-886c-4ff7a903e705/knight_corpse_hades_style_1772400516255.png"
    }
    
    output_dir = "assets/images/characters/main"
    
    for filename, source_path in sources.items():
        output_path = os.path.join(output_dir, filename)
        if os.path.exists(source_path):
            scale_image(source_path, output_path, scale_factor=0.30)
        else:
            print(f"Warning: source {source_path} not found.")
