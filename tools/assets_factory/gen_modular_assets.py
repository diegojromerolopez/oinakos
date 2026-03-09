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

print("Generating Master Human...")
macro_details = TargetService.get_default_macro_info_dict()
macro_details["gender"] = 1.0 # Male
macro_details["age"] = 0.5    # Young adult
macro_details["muscle"] = 0.8 # Stronger
macro_details["weight"] = 0.6
macro_details["race"]["caucasian"] = 1.0

basemesh = HumanService.create_human(
    mask_helpers=True,
    feet_on_ground=True, 
    scale=0.15, # Heroic scale
    macro_detail_dict=macro_details
)

# --- PROCEDURAL CLOTHING GENERATION ---
# Instead of libraries, we "extrude" clothes from the body mesh
def create_piece(name, color, shrink_val=0.015):
    # Duplicate basemesh to create clothing
    bpy.ops.object.select_all(action='DESELECT')
    basemesh.select_set(True)
    bpy.context.view_layer.objects.active = basemesh
    bpy.ops.object.duplicate()
    piece = bpy.context.active_object
    piece.name = name
    
    # Scale up slightly (Inflate)
    bpy.ops.object.mode_set(mode='EDIT')
    bpy.ops.mesh.select_all(action='SELECT')
    bpy.ops.transform.shrink_fatten(value=shrink_val)
    bpy.ops.object.mode_set(mode='OBJECT')
    
    # Material
    mat = bpy.data.materials.new(name=f"Mat_{name}")
    mat.use_nodes = True
    mat.node_tree.nodes["Principled BSDF"].inputs["Base Color"].default_value = color
    piece.data.materials.clear()
    piece.data.materials.append(mat)
    return piece

print("Creating Outfit Layers...")
tunic = create_piece("Layer_Tunic", (0.1, 0.2, 0.5, 1.0), 0.02)
leather = create_piece("Layer_LeatherArmor", (0.2, 0.1, 0.05, 1.0), 0.04)

# Create Face Variants via Targets
# We'll just do 2 distinct faces for now
faces = []
# Face 1 is the default
# Face 2: Old man (we'll modify a copy later)

# Setup Scene
import mathutils

def get_bounding_box(obj):
    local_coords = obj.bound_box
    world_coords = [obj.matrix_world @ mathutils.Vector(coord) for coord in local_coords]
    min_z = min(c.z for c in world_coords)
    max_z = max(c.z for c in world_coords)
    return (min_z + max_z) / 2.0, max_z - min_z

center_z, height = get_bounding_box(basemesh)

bpy.ops.object.camera_add()
cam = bpy.context.active_object
bpy.context.scene.camera = cam
cam.data.type = 'ORTHO'
cam.data.ortho_scale = height * 1.5 
cam.location = (10, -10, 10 + center_z)

bpy.ops.object.empty_add(type='PLAIN_AXES', location=(0, 0, center_z))
target = bpy.context.active_object
constr = cam.constraints.new(type='TRACK_TO')
constr.target = target
constr.track_axis = 'TRACK_NEGATIVE_Z'
constr.up_axis = 'UP_Y'

bpy.context.scene.render.resolution_x = 160
bpy.context.scene.render.resolution_y = 160
bpy.context.scene.render.film_transparent = True

# Lighting
bpy.ops.object.light_add(type='SUN', location=(5, -5, 10))
bpy.context.active_object.data.energy = 5
bpy.context.active_object.rotation_euler = (math.radians(45), math.radians(15), math.radians(45))

# --- MASS RENDERING ---
# 1. Create a rotation handle for everyone
bpy.ops.object.empty_add(type='PLAIN_AXES', location=(0,0,0))
handle = bpy.context.active_object
basemesh.parent = handle
tunic.parent = handle
leather.parent = handle

DIRECTIONS = {"n": 0, "ne": 45, "e": 90, "se": 135, "s": 180, "sw": 225, "w": 270, "nw": 315}

def render_collection(obj_list, folder_name):
    # Hide all
    basemesh.hide_render = True
    tunic.hide_render = True
    leather.hide_render = True
    
    for o in obj_list: o.hide_render = False
    
    base_path = f"/Users/diegoj/repos/oinakos/assets/images/characters/human_male/modular/{folder_name}"
    os.makedirs(base_path, exist_ok=True)
    
    for dir_name, angle in DIRECTIONS.items():
        handle.rotation_euler.z = math.radians(angle)
        bpy.context.scene.render.filepath = os.path.join(base_path, f"static_{dir_name}.png")
        bpy.ops.render.render(write_still=True)

print("Starting Multi-Layer Render...")
render_collection([basemesh], "body")
render_collection([tunic], "tunic_blue")
render_collection([leather], "armor_leather")

# Save file
bpy.ops.wm.save_as_mainfile(filepath="/Users/diegoj/repos/oinakos/tools/assets_factory/modular_human_master.blend")
print("DONE! Created body, tunic, and armor layers.")
