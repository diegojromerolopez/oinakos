import sys
import os
import subprocess

def generate_audio(text, model_path, output_file):
    """
    Generates audio using piper-tts.
    """
    print(f"Generating: '{text}' -> {output_file} (using {model_path})")
    
    # Ensure output directory exists
    os.makedirs(os.path.dirname(output_file), exist_ok=True)
    
    # Piper CLI usage: echo "text" | python -m piper --model model.onnx --output_file output.wav
    try:
        if output_file.endswith(".mp3"):
            temp_wav = output_file + ".wav"
            # Piper only outputs WAV, so we generate a temp WAV and convert
            process = subprocess.Popen(
                [sys.executable, "-m", "piper", "--model", model_path, "--output-file", temp_wav],
                stdin=subprocess.PIPE, text=True
            )
            process.communicate(input=text)
            if process.returncode == 0:
                subprocess.run(["ffmpeg", "-y", "-i", temp_wav, "-codec:a", "libmp3lame", "-qscale:a", "2", output_file], check=True)
                os.remove(temp_wav)
                return True
            return False

        process = subprocess.Popen(
            [sys.executable, "-m", "piper", "--model", model_path, "--output-file", output_file],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        stdout, stderr = process.communicate(input=text)
        
        if process.returncode != 0:
            print(f"Error generating audio: {stderr}")
            return False
            
        return True
    except Exception as e:
        print(f"Exception: {e}")
        return False

if __name__ == "__main__":
    if len(sys.argv) < 4:
        print("Usage: python piper_gen.py <text> <model_path> <output_file>")
        sys.exit(1)
        
    text = sys.argv[1]
    model = sys.argv[2]
    output = sys.argv[3]
    
    if generate_audio(text, model, output):
        print("Success")
    else:
        print("Failed")
        sys.exit(1)
