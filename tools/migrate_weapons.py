import os
import glob
import yaml

def stringify_name(name):
    return name.lower().replace(" ", "_")

def main():
    search_dirs = ["data/characters", "data/npcs", "data/archetypes"]
    files_to_process = []
    
    for d in search_dirs:
        for root, _, files in os.walk(d):
            for file in files:
                if file.endswith(".yaml") or file.endswith(".yml"):
                    files_to_process.append(os.path.join(root, file))

    unique_weapons = {}

    for filepath in files_to_process:
        with open(filepath, 'r') as f:
            try:
                data = yaml.safe_load(f)
            except:
                continue
                
        if not data:
            continue
            
        if "weapon" in data and isinstance(data["weapon"], dict):
            w = data["weapon"]
            w_name = w.get("name", "Unknown Weapon")
            w_id = stringify_name(w_name)
            
            # Save to unique weapons to create
            if w_id not in unique_weapons:
                # Build object config
                obj = {
                    "name": w_name,
                    "description": f"A standard {w_name.lower()}.",
                    "weight": 1.0,
                    "type": "weapon",
                    "slot": "weapon",
                    "value": 50,
                    "combat": {
                        "name": w_name,
                        "type": w.get("type", "melee"),
                        "damage": w.get("damage", {"min":1, "max":2})
                    }
                }
                if "max_distance" in w:
                    obj["combat"]["max_distance"] = w["max_distance"]
                unique_weapons[w_id] = obj

            # Update entity config
            del data["weapon"]
            if "equipment" not in data:
                data["equipment"] = {}
            if isinstance(data["equipment"], list):
                # Should not happen if previous script worked, but just in case
                new_eq = {}
                for x in data["equipment"]:
                    new_eq["weapon"] = x # guess
                data["equipment"] = new_eq
                
            data["equipment"]["weapon"] = w_id
            if "inventory" not in data:
                data["inventory"] = []
                
            # Keep yaml order somewhat sane by dumping back
            with open(filepath, 'w') as f:
                yaml.dump(data, f, sort_keys=False)
                
            print(f"Updated {filepath} -> {w_id}")

    # Write out the weapon objects
    for w_id, obj in unique_weapons.items():
        obj_path = f"data/objects/{w_id}.yaml"
        if not os.path.exists(obj_path):
            with open(obj_path, 'w') as f:
                yaml.dump(obj, f, sort_keys=False)
            print(f"Created weapon object: {obj_path}")

if __name__ == "__main__":
    main()
