import sys
from diffusers import StableDiffusionPipeline
import torch

def generate_image(prompt, output_path):
    pipe = StableDiffusionPipeline.from_pretrained("CompVis/stable-diffusion-v1-4") # Use a fast model
    if torch.backends.mps.is_available():
        pipe = pipe.to("mps")
    elif torch.cuda.is_available():
        pipe = pipe.to("cuda")
    
    # Generate image
    image = pipe(prompt, num_inference_steps=20).images[0]
    image.save(output_path)
    print(f"Generated image: {output_path}")

if __name__ == "__main__":
    if len(sys.argv) < 5:
        print("Usage: python generate_assets.py --prompt '<prompt>' --output <path>")
        sys.exit(1)
    
    prompt = sys.argv[2]
    output = sys.argv[4]
    generate_image(prompt, output)
