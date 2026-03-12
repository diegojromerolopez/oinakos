import os
import yaml
import subprocess
import sys
from pathlib import Path

# Phrases mapping
PHRASES = {
    "generic": {
        "hit": "Ugh!",
        "death": "I fall...",
        "attack": [
            "You shall not pass!",
            "I will end you!",
            "Leave this place!",
            "For the crown!",
            "Die, interloper!"
        ]
    },
    "orc": {
        "hit": "Gah!",
        "death": "Blood and thunder!",
        "attack": [
            "Meat's back on the menu!",
            "I'll crush your bones!",
            "For the horde!",
            "You're a small one!",
            "Weak human!"
        ]
    },
    "peasant": {
        "hit": "Ouch!",
        "death": "My farm... my poor farm...",
        "attack": [
            "Get out of my field!",
            "You are trespassing!",
            "For the village!",
            "Leave us alone!",
            "I will defend my home!"
        ]
    },
    "demon": {
        "hit": "Pah!",
        "death": "I return to the abyss...",
        "attack": [
            "Your soul is mine!",
            "Burn in eternal fire!",
            "Darkness consumes you!",
            "Feeble mortal!",
            "The end is nigh!"
        ]
    },
    "giant": {
        "hit": "Boom!",
        "death": "The mountain falls...",
        "attack": [
            "I'll step on you!",
            "Puny knight!",
            "Fee-fi-fo-fum!",
            "Crush!",
            "You are but a bug!"
        ]
    },
    "goblin": {
        "hit": "Eeeek!",
        "death": "Not fair!",
        "attack": [
            "Shiny gold!",
            "Stab stab stab!",
            "We surround you!",
            "Running won't help!",
            "Hehe, gotcha!"
        ]
    },
    "magi": {
        "hit": "Focus lost!",
        "death": "The light fades...",
        "attack": [
            "Arcane power!",
            "You lack the spark!",
            "Knowledge is power!",
            "Witness the truth!",
            "A simple spell..."
        ]
    }
}

MODEL_PATH_FEMALE = "models/en_US-lessac-medium.onnx"
MODEL_PATH_MALE = "models/en_US-ryan-medium.onnx"

def get_phrases(category):
    category = category.lower()
    for key in PHRASES:
        if key in category:
            return PHRASES[key]
    return PHRASES["generic"]

def generate_entity_audio(entity_id, category, sub_path, gender="female", force=False):
    # Determine base audio path
    base_audio_dir = Path("assets/audio") / sub_path
    base_audio_dir.mkdir(parents=True, exist_ok=True)
    
    phrases = get_phrases(category)
    if gender == "male":
        model = MODEL_PATH_MALE
    elif gender == "none":
        model = MODEL_PATH_FEMALE # Use Lessac as default neutral for now
    else:
        model = MODEL_PATH_FEMALE
    
    # hit.mp3
    generate(phrases["hit"], base_audio_dir / "hit.mp3", model, force)
    
    # death.mp3
    generate(phrases["death"], base_audio_dir / "death.mp3", model, force)
    
    # attack_1..5.mp3
    for i, line in enumerate(phrases["attack"]):
        if i >= 5: break
        generate(line, base_audio_dir / f"attack_{i+1}.mp3", model, force)

def generate(text, output_file, model_path, force=False):
    if output_file.exists() and not force:
        return
    
    print(f"Generating {output_file} (gender model: {model_path}): {text}")
    subprocess.run([
        sys.executable, "scripts/piper_gen.py", 
        text, model_path, str(output_file)
    ], check=True)

def main():
    # Entities to force re-generation for (e.g. if we changed the model or gender)
    force_ids = []

    # 1. Process Archetypes
    arch_root = Path("data/archetypes")
    arch_cache = {} # id -> gender for NPC inheritance
    for yaml_file in arch_root.rglob("*.yaml"):
        with open(yaml_file, 'r') as f:
            data = yaml.safe_load(f)
        
        # Priority: YAML gender attribute
        gender = data.get("gender", "female")
        arch_cache[data.get("id")] = gender
            
        rel_path = yaml_file.relative_to(arch_root)
        category = rel_path.parts[0]
        sub_path = Path("archetypes") / rel_path.with_suffix("")
        
        force = yaml_file.stem in force_ids or data.get("id") in force_ids
        generate_entity_audio(yaml_file.stem, category, sub_path, gender, force)

    # 2. Process Characters
    char_root = Path("data/characters")
    for yaml_file in char_root.rglob("*.yaml"):
        with open(yaml_file, 'r') as f:
            data = yaml.safe_load(f)
        
        gender = data.get("gender", "male") # Oinakos is male
        sub_path = Path("characters") / yaml_file.stem
        force = yaml_file.stem in force_ids or data.get("id") in force_ids
        generate_entity_audio(yaml_file.stem, "hero", sub_path, gender, force)

    # 3. Process NPCs (Unique or Profiles)
    npc_root = Path("data/npcs")
    for yaml_file in npc_root.rglob("*.yaml"):
        with open(yaml_file, 'r') as f:
            data = yaml.safe_load(f)
        
        # Priority: YAML gender attribute > Archetype gender
        gender = data.get("gender")
        if not gender and data.get("archetype"):
            gender = arch_cache.get(data["archetype"])
        
        if not gender:
            gender = "female" # Default fallback
        
        category = "unique" if data.get("unique") else "generic"
        if "guard" in yaml_file.name.lower():
            category = "generic" # Use generic phrases for guards
            
        sub_path = Path("npcs") / yaml_file.stem
        force = yaml_file.stem in force_ids or data.get("id") in force_ids
        generate_entity_audio(yaml_file.stem, category, sub_path, gender, force)

if __name__ == "__main__":
    main()
