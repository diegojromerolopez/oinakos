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
from bl_ext.blender_org.mpfb.services import HumanService, TargetService, LogService

def clear_the_scene():
    bpy.ops.object.select_all(action='SELECT')
    bpy.ops.object.delete()
    for mesh in bpy.data.meshes: bpy.data.meshes.remove(mesh)
    for mat in bpy.data.materials: bpy.data.materials.remove(mat)

def setup_camera_zoom_head(height, center_z):
    bpy.ops.object.camera_add()
    cam = bpy.context.active_object
    bpy.context.scene.camera = cam
    cam.data.type = 'ORTHO'
    cam.data.ortho_scale = height * 1.5 # Same zoom as body to maintain alignment
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
    return cam

def generate_face(age_val, race_asian, name):
    clear_the_scene()
    macro_details = TargetService.get_default_macro_info_dict()
    macro_details["gender"] = 1.0 
    macro_details["age"] = age_val
    macro_details["race"]["caucasian"] = 1.0 - race_asian
    macro_details["race"]["asian"] = race_asian
    
    basemesh = HumanService.create_human(
        mask_helpers=True,
        feet_on_ground=True, 
        scale=0.15, 
        macro_detail_dict=macro_details
    )
    
    # Mask everything EXECPT the head
    # In a real game we would use vertex groups, but for this automated sprites
    # we'll just use a shader that hides everything below Z=1.5
    mat = bpy.data.materials.new(name="FaceMask")
    mat.use_nodes = True
    nodes = mat.node_tree.nodes
    links = mat.node_tree.links
    output = nodes.get("Material Output")
    bsdf = nodes.get("Principled BSDF")
    
    coord = nodes.new("ShaderNodeTexCoord")
    sep = nodes.new("ShaderNodeSeparateColor") # Use Separate Color or Separate XYZ
    # In Blender 4.0+ Separate XYZ is often available
    sep_xyz = nodes.new("ShaderNodeSeparateXYZ")
    math_node = nodes.new("ShaderNodeMath")
    math_node.operation = 'GREATER_THAN'
    math_node.inputs[1].default_value = 1.55 # Height of head
    
    mix = nodes.new("ShaderNodeMixShader")
    transp = nodes.new("ShaderNodeBsdfTransparent")
    
    links.new(coord.outputs["Generated"], sep_xyz.inputs["Vector"])
    links.new(sep_xyz.outputs["Z"], math_node.inputs[0])
    links.new(math_node.outputs["Value"], mix.inputs["Factor"])
    links.new(transp.outputs["BSDF"], mix.inputs[1])
    links.new(bsdf.outputs["BSDF"], mix.inputs[2])
    links.new(mix.outputs["Shader"], output.inputs["Surface"])
    
    basemesh.data.materials.clear()
    basemesh.data.materials.append(mat)
    
    # Setup rendering
    center_z = 1.3391505202744156 # From previous run
    height = 2.6783011513762176
    setup_camera_zoom_head(height, center_z)
    
    bpy.ops.object.empty_add(type='PLAIN_AXES', location=(0,0,0))
    handle = bpy.context.active_object
    basemesh.parent = handle
    
    DIRECTIONS = {"n": 0, "ne": 45, "e": 90, "se": 135, "s": 180, "sw": 225, "w": 270, "nw": 315}
    base_path = f"/Users/diegoj/repos/oinakos/assets/images/characters/human_male/modular/face_{name}"
    os.makedirs(base_path, exist_ok=True)
    
    for dir_name, angle in DIRECTIONS.items():
        handle.rotation_euler.z = math.radians(angle)
        bpy.context.scene.render.filepath = os.path.join(base_path, f"static_{dir_name}.png")
        bpy.ops.render.render(write_still=True)

print("Generating Face Variants...")
generate_face(0.5, 0.0, "young") # Standard Young
generate_face(1.0, 0.0, "old")   # Old Man
generate_face(0.5, 0.8, "asian") # Different Race Features

print("DONE! Created 3 face variants.")
