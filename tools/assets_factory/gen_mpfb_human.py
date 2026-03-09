import sys
import os
import bpy
import math

# Add MPFB to path
ext_base = "/Users/diegoj/Library/Application Support/Blender/5.0/extensions/"
sys.path.append(ext_base)

# Prepare MPFB Contextual Information
import bl_ext.blender_org.mpfb as mpfb
mpfb.MPFB_CONTEXTUAL_INFORMATION = {
    "__package__": "bl_ext.blender_org.mpfb",
    "__package_short__": "mpfb",
    "__file__": os.path.join(ext_base, "blender_org/mpfb/__init__.py")
}

# Mock get_preference
def mock_get_pref(name):
    if name == "mpfb_user_data": return os.path.expanduser("~/Documents/MakeHuman/v1py3")
    return None
mpfb.get_preference = mock_get_pref

# Import services
from bl_ext.blender_org.mpfb.services import HumanService, TargetService

def clear_the_scene():
    bpy.ops.object.select_all(action='SELECT')
    bpy.ops.object.delete()
    for mesh in bpy.data.meshes: bpy.data.meshes.remove(mesh)
    for mat in bpy.data.materials: bpy.data.materials.remove(mat)

clear_the_scene()

print("Generating Human Male...")
macro_details = TargetService.get_default_macro_info_dict()
macro_details["muscle"] = 0.8
macro_details["weight"] = 0.7
macro_details["height"] = 0.5
macro_details["gender"] = 1.0 # Male
macro_details["age"] = 0.5    # Young adult
macro_details["race"]["caucasian"] = 1.0

basemesh = HumanService.create_human(
    mask_helpers=True,
    feet_on_ground=True, 
    scale=0.15, # 50% larger than before (approx 2.67m tall)
    macro_detail_dict=macro_details
)

import mathutils

def get_bounding_box(obj):
    """Calculates world-space bounding box."""
    local_coords = obj.bound_box
    world_coords = [obj.matrix_world @ mathutils.Vector(coord) for coord in local_coords]
    
    min_x = min(c.x for c in world_coords)
    max_x = max(c.x for c in world_coords)
    min_y = min(c.y for c in world_coords)
    max_y = max(c.y for c in world_coords)
    min_z = min(c.z for c in world_coords)
    max_z = max(c.z for c in world_coords)
    
    return (min_x, max_x), (min_y, max_y), (min_z, max_z)

# Setup Scene
def setup_oinakos_scene(obj):
    # Find center of character
    (min_x, max_x), (min_y, max_y), (min_z, max_z) = get_bounding_box(obj)
    center_z = (min_z + max_z) / 2.0
    height = max_z - min_z
    
    print(f"Character Height: {height}, Center Z: {center_z}")
    
    # Create target for camera
    bpy.ops.object.empty_add(type='PLAIN_AXES', location=(0, 0, center_z))
    target = bpy.context.active_object
    target.name = "CamTarget"

    bpy.ops.object.camera_add()
    cam = bpy.context.active_object
    bpy.context.scene.camera = cam
    cam.data.type = 'ORTHO'
    # Ortho scale usually needs to be slightly larger than height to allow for isometric diagonal
    cam.data.ortho_scale = height * 2.5 # Very safe zoom out
    
    # Distance doesn't matter for Ortho sizing, but helps for clipping
    dist = 15
    cam.location = (dist, -dist, dist + center_z) 
    
    # Look at the target
    constr = cam.constraints.new(type='TRACK_TO')
    constr.target = target
    constr.track_axis = 'TRACK_NEGATIVE_Z'
    constr.up_axis = 'UP_Y'
    
    bpy.context.scene.render.resolution_x = 160
    bpy.context.scene.render.resolution_y = 160
    bpy.context.scene.render.film_transparent = True
    
    # Setup lighting
    bpy.ops.object.light_add(type='SUN', location=(5, -5, 10))
    sun = bpy.context.active_object
    sun.data.energy = 5
    sun.rotation_euler = (math.radians(45), math.radians(15), math.radians(45))

setup_oinakos_scene(basemesh)

# --- RENDER ALL 8 DIRECTIONS ---
# Use a rotation handle
bpy.ops.object.select_all(action='DESELECT')
basemesh.select_set(True)
bpy.ops.object.empty_add(type='PLAIN_AXES', location=(0,0,0))
handle = bpy.context.active_object
basemesh.parent = handle

DIRECTIONS = {
    "n": 0,
    "ne": 45,
    "e": 90,
    "se": 135,
    "s": 180,
    "sw": 225,
    "w": 270,
    "nw": 315
}

BASE_DIR = "/Users/diegoj/repos/oinakos/assets/images/characters/human_male/mpfb"
os.makedirs(BASE_DIR, exist_ok=True)

for dir_name, angle in DIRECTIONS.items():
    handle.rotation_euler.z = math.radians(angle)
    path = os.path.join(BASE_DIR, f"static_{dir_name}.png")
    bpy.context.scene.render.filepath = path
    bpy.ops.render.render(write_still=True)
    print(f"Saved {dir_name} to {path}")

# Final .blend save
bpy.ops.wm.save_as_mainfile(filepath="/Users/diegoj/repos/oinakos/tools/assets_factory/mpfb_human_full_set.blend")
