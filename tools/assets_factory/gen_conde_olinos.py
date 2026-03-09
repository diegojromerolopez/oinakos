import sys
import os
import bpy
import math
import mathutils

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

def mock_get_pref(name):
    if name == "mpfb_user_data": return os.path.expanduser("~/Documents/MakeHuman/v1py3")
    return None
mpfb.get_preference = mock_get_pref

from bl_ext.blender_org.mpfb.services import HumanService, TargetService

def clear_the_scene():
    bpy.ops.object.select_all(action='SELECT')
    bpy.ops.object.delete()
    for mesh in bpy.data.meshes: bpy.data.meshes.remove(mesh)
    for mat in bpy.data.materials: bpy.data.materials.remove(mat)

clear_the_scene()

print("Generating Master Human with Rig...")
macro_details = TargetService.get_default_macro_info_dict()
macro_details["gender"] = 1.0 
macro_details["age"] = 0.5
macro_details["muscle"] = 0.6 
macro_details["weight"] = 0.5
macro_details["race"]["caucasian"] = 1.0

basemesh = HumanService.create_human(
    mask_helpers=True,
    feet_on_ground=True, 
    scale=0.12, # Slightly larger
    macro_detail_dict=macro_details
)

# Add Rig
HumanService.add_builtin_rig(basemesh, "game_engine")
rig = bpy.context.active_object
bpy.context.view_layer.objects.active = rig
bpy.ops.object.mode_set(mode='POSE')

# --- MATERIALS ---
skin_mat = bpy.data.materials.new(name="SienaWhiteSkin")
skin_mat.use_nodes = True
skin_mat.node_tree.nodes["Principled BSDF"].inputs["Base Color"].default_value = (0.85, 0.75, 0.65, 1.0)
basemesh.data.materials.clear()
basemesh.data.materials.append(skin_mat)

PRIMARY_COLOR = (1.0, 0.0, 1.0, 1.0)   # Magenta
SECONDARY_COLOR = (1.0, 1.0, 0.0, 1.0) # Yellow

# --- CLOTHING / ARMOR SHELLS ---
def create_skin(name, color, shrink_val, metallic=0.0):
    bpy.ops.object.mode_set(mode='OBJECT')
    bpy.ops.object.select_all(action='DESELECT')
    basemesh.select_set(True)
    bpy.context.view_layer.objects.active = basemesh
    bpy.ops.object.duplicate()
    piece = bpy.context.active_object
    piece.name = name
    bpy.ops.object.mode_set(mode='EDIT')
    bpy.ops.mesh.select_all(action='SELECT')
    bpy.ops.transform.shrink_fatten(value=shrink_val)
    bpy.ops.object.mode_set(mode='OBJECT')
    mat = bpy.data.materials.new(name=f"Mat_{name}")
    mat.use_nodes = True
    mat.node_tree.nodes["Principled BSDF"].inputs["Base Color"].default_value = color
    mat.node_tree.nodes["Principled BSDF"].inputs["Metallic"].default_value = metallic
    piece.data.materials.clear()
    piece.data.materials.append(mat)
    piece.parent = rig
    mod = piece.modifiers.new(name="Armature", type='ARMATURE')
    mod.object = rig
    bpy.context.view_layer.objects.active = rig
    bpy.ops.object.mode_set(mode='POSE')
    return piece

armor = create_skin("Armor", (0.5, 0.5, 0.55, 1.0), 0.03, metallic=0.9)
hair_beard = create_skin("HeadDetails", (0.01, 0.01, 0.01, 1.0), 0.015)
tunic = create_skin("Tunic", PRIMARY_COLOR, 0.02)

def create_cape():
    bpy.ops.object.mode_set(mode='OBJECT')
    bpy.ops.mesh.primitive_plane_add(size=1.0)
    cape = bpy.context.active_object
    cape.name = "Cape"
    cape.scale = (0.4, 0.8, 1.0)
    cape.location = (0, 0.15, 1.1)
    cape.rotation_euler = (math.radians(80), 0, 0)
    mat = bpy.data.materials.new(name="Mat_Cape")
    mat.use_nodes = True
    mat.node_tree.nodes["Principled BSDF"].inputs["Base Color"].default_value = SECONDARY_COLOR
    cape.data.materials.clear()
    cape.data.materials.append(mat)
    cape.parent = rig
    cape.parent_type = 'BONE'
    cape.parent_bone = "spine_03"
    bpy.context.view_layer.objects.active = rig
    bpy.ops.object.mode_set(mode='POSE')
    return cape

cape = create_cape()

def create_sword():
    bpy.ops.object.mode_set(mode='OBJECT')
    bpy.ops.mesh.primitive_cube_add(size=1.0)
    blade = bpy.context.active_object
    blade.scale = (0.02, 0.005, 0.6)
    blade.location = (0, 0, 0.7)
    bpy.ops.mesh.primitive_cube_add(size=1.0)
    guard = bpy.context.active_object
    guard.scale = (0.2, 0.02, 0.02)
    guard.location = (0, 0, 0.1)
    bpy.ops.mesh.primitive_cylinder_add(radius=0.015, depth=0.2)
    handle = bpy.context.active_object
    handle.location = (0, 0, 0)
    bpy.ops.object.select_all(action='DESELECT')
    blade.select_set(True)
    guard.select_set(True)
    handle.select_set(True)
    bpy.context.view_layer.objects.active = handle
    bpy.ops.object.join()
    handle.parent = rig
    handle.parent_type = 'BONE'
    handle.parent_bone = "hand_r"
    handle.location = (0, 0, 0)
    handle.rotation_euler = (math.radians(-90), 0, math.radians(90))
    mat = bpy.data.materials.new(name="SwordMetal")
    mat.use_nodes = True
    mat.node_tree.nodes["Principled BSDF"].inputs["Base Color"].default_value = (0.8, 0.8, 0.85, 1.0)
    mat.node_tree.nodes["Principled BSDF"].inputs["Metallic"].default_value = 1.0
    handle.data.materials.clear()
    handle.data.materials.append(mat)
    bpy.context.view_layer.objects.active = rig
    bpy.ops.object.mode_set(mode='POSE')
    return handle

sword = create_sword()

# --- POSING ACTIONS ---
def set_pose_static():
    for bone in rig.pose.bones:
        bone.rotation_mode = 'XYZ'
        bone.rotation_euler = (0, 0, 0)
        bone.location = (0, 0, 0)
    # Strict vertical arms: Fingers pointing to the ground
    # In MPFB game_engine rig, ~90 deg is vertical down
    if "upperarm_l" in rig.pose.bones: rig.pose.bones["upperarm_l"].rotation_euler = (0, 0, math.radians(90))
    if "upperarm_r" in rig.pose.bones: rig.pose.bones["upperarm_r"].rotation_euler = (0, 0, math.radians(-90))
    # No elbow bend, keeping arm perfectly straight
    if "lowerarm_l" in rig.pose.bones: rig.pose.bones["lowerarm_l"].rotation_euler = (0, 0, 0)
    if "lowerarm_r" in rig.pose.bones: rig.pose.bones["lowerarm_r"].rotation_euler = (0, 0, 0)
    if "hand_l" in rig.pose.bones: rig.pose.bones["hand_l"].rotation_euler = (0, 0, 0)
    if "hand_r" in rig.pose.bones: rig.pose.bones["hand_r"].rotation_euler = (0, 0, 0)

def set_pose_walk_cycle(frame_idx):
    set_pose_static()
    # 4-frame walk cycle: 0: L-fwd, 1: Passing, 2: R-fwd, 3: Passing
    sweep = math.radians(35)
    arm_sweep = math.radians(10) # Subtle arm movement
    bob_down = -0.05
    bob_up = 0.0
    
    if "Root" in rig.pose.bones:
        rig.pose.bones["Root"].location.z = bob_down if frame_idx % 2 == 0 else bob_up

    if frame_idx == 0: # Left leg forward
        if "thigh_l" in rig.pose.bones: 
            rig.pose.bones["thigh_l"].rotation_euler = (sweep, 0, 0)
            rig.pose.bones["calf_l"].rotation_euler = (math.radians(-30), 0, 0)
        if "thigh_r" in rig.pose.bones: 
            rig.pose.bones["thigh_r"].rotation_euler = (-sweep * 0.8, 0, 0)
            rig.pose.bones["calf_r"].rotation_euler = (math.radians(-15), 0, 0)
        # Subtle arm swing (Inverse of legs)
        if "upperarm_r" in rig.pose.bones: rig.pose.bones["upperarm_r"].rotation_euler.x = arm_sweep
        if "upperarm_l" in rig.pose.bones: rig.pose.bones["upperarm_l"].rotation_euler.x = -arm_sweep
    elif frame_idx == 1: # Passing
        if "thigh_l" in rig.pose.bones: rig.pose.bones["thigh_l"].rotation_euler = (math.radians(5), 0, 0)
        if "thigh_r" in rig.pose.bones: rig.pose.bones["thigh_r"].rotation_euler = (math.radians(-5), 0, 0)
    elif frame_idx == 2: # Right leg forward
        if "thigh_r" in rig.pose.bones: 
            rig.pose.bones["thigh_r"].rotation_euler = (sweep, 0, 0)
            rig.pose.bones["calf_r"].rotation_euler = (math.radians(-30), 0, 0)
        if "thigh_l" in rig.pose.bones: 
            rig.pose.bones["thigh_l"].rotation_euler = (-sweep * 0.8, 0, 0)
            rig.pose.bones["calf_l"].rotation_euler = (math.radians(-15), 0, 0)
        # Subtle arm swing (Inverse of legs)
        if "upperarm_l" in rig.pose.bones: rig.pose.bones["upperarm_l"].rotation_euler.x = arm_sweep
        if "upperarm_r" in rig.pose.bones: rig.pose.bones["upperarm_r"].rotation_euler.x = -arm_sweep
    elif frame_idx == 3: # Passing
        if "thigh_r" in rig.pose.bones: rig.pose.bones["thigh_r"].rotation_euler = (math.radians(5), 0, 0)
        if "thigh_l" in rig.pose.bones: rig.pose.bones["thigh_l"].rotation_euler = (math.radians(-5), 0, 0)

def set_pose_walk1(): set_pose_walk_cycle(0)
def set_pose_walk2(): set_pose_walk_cycle(1)
def set_pose_walk3(): set_pose_walk_cycle(2)
def set_pose_walk4(): set_pose_walk_cycle(3)

def set_pose_attack():
    set_pose_static()
    # Professional Lunge
    if "thigh_l" in rig.pose.bones: rig.pose.bones["thigh_l"].rotation_euler = (math.radians(45), 0, 0)
    if "thigh_r" in rig.pose.bones: rig.pose.bones["thigh_r"].rotation_euler = (math.radians(-25), 0, 0)
    if "upperarm_r" in rig.pose.bones: rig.pose.bones["upperarm_r"].rotation_euler = (math.radians(70), math.radians(-40), math.radians(-30))
    if "lowerarm_r" in rig.pose.bones: rig.pose.bones["lowerarm_r"].rotation_euler = (0, 0, math.radians(-20))

def set_pose_hit():
    set_pose_static()
    # Snap back
    if "spine_01" in rig.pose.bones: rig.pose.bones["spine_01"].rotation_euler = (math.radians(-15), 0, 0)
    if "head" in rig.pose.bones: rig.pose.bones["head"].rotation_euler = (math.radians(15), 0, 0)

# --- CAMERA ---
def setup_camera():
    bpy.ops.object.mode_set(mode='OBJECT')
    bpy.ops.object.camera_add()
    cam = bpy.context.active_object
    bpy.context.scene.camera = cam
    cam.data.type = 'ORTHO'
    cam.data.ortho_scale = 3.2
    cam.location = (10, -10, 10 + 1.1)
    bpy.ops.object.empty_add(type='PLAIN_AXES', location=(0, 0, 1.1))
    target = bpy.context.active_object
    constr = cam.constraints.new(type='TRACK_TO')
    constr.target = target
    constr.track_axis = 'TRACK_NEGATIVE_Z'
    constr.up_axis = 'UP_Y'
    bpy.context.scene.render.resolution_x = 160
    bpy.context.scene.render.resolution_y = 160
    bpy.context.scene.render.film_transparent = True
    bpy.ops.object.light_add(type='SUN', location=(5, -5, 10))
    bpy.context.active_object.data.energy = 5
    bpy.context.view_layer.objects.active = rig
    bpy.ops.object.mode_set(mode='POSE')

setup_camera()

# --- RENDERING ---
# Correct Mapping for Isometric View Profile looking:
# 315 deg: faces camera (South). 45 deg: faces Right (Profile East). 135 deg: faces Back (North).
DIRECTIONS = {"s": 315, "se": 0, "e": 45, "ne": 90, "n": 135}
ACTIONS = {"static": set_pose_static, "walk1": set_pose_walk1, "walk2": set_pose_walk2, "walk3": set_pose_walk3, "walk4": set_pose_walk4, "attack": set_pose_attack, "hit": set_pose_hit}
LAYERS = {
    "body": [basemesh],
    "head_details": [hair_beard],
    "armor": [armor],
    "tunic": [tunic],
    "cape": [cape],
    "weapon_r": [sword]
}

print("Starting PRODUCTION (Optimized for Flipping)...")
for action_name, pose_func in ACTIONS.items():
    pose_func()
    for layer_name, objs in LAYERS.items():
        # Ensure only the current layer and dependencies are visible
        # For clothes, we hide the basemesh bones if they poke through?
        # Actually, shrink_fatten should handle it.
        for o in [basemesh, armor, hair_beard, tunic, cape, sword]: o.hide_render = True
        for o in objs: o.hide_render = False

        base_path = f"/Users/diegoj/repos/oinakos/assets/images/characters/conde_olinos/paperdoll/{layer_name}"
        os.makedirs(base_path, exist_ok=True)
        for dir_name, angle in DIRECTIONS.items():
            rig.rotation_euler.z = math.radians(angle)
            bpy.context.scene.render.filepath = os.path.join(base_path, f"{action_name}_{dir_name}.png")
            bpy.ops.render.render(write_still=True)

bpy.ops.wm.save_as_mainfile(filepath="/Users/diegoj/repos/oinakos/tools/assets_factory/conde_olinos_master.blend")
print("PRODUCTION COMPLETE!")
