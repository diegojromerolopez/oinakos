import bpy
import math
import os

# Configuration
BASE_DIR = "/Users/diegoj/repos/oinakos/assets/images/characters/human_male"
os.makedirs(os.path.join(BASE_DIR, "body"), exist_ok=True)
os.makedirs(os.path.join(BASE_DIR, "hair_brown"), exist_ok=True)
os.makedirs(os.path.join(BASE_DIR, "tunic_blue"), exist_ok=True)

def cleanup():
    bpy.ops.wm.read_factory_settings(use_empty=True)

def create_realistic_human():
    # 1. Create Body Collection
    body_col = bpy.data.collections.new("Body")
    bpy.context.scene.collection.children.link(body_col)
    
    # We use METABALLS for organic "melting" skin
    mball = bpy.data.metaballs.new("HumanBody")
    obj = bpy.data.objects.new("HumanBody", mball)
    body_col.objects.link(obj)
    
    # Settings for smoothness
    mball.resolution = 0.05
    mball.render_resolution = 0.02

    def add_limb(loc, size, name):
        ele = mball.elements.new()
        ele.co = loc
        ele.radius = size
        return ele

    # TORSO (V-shape)
    add_limb((0, 0, 1.4), 0.25, "Chest")
    add_limb((0, 0, 1.1), 0.2, "Waist")
    add_limb((0, 0, 0.9), 0.22, "Hips")
    
    # ARMS (Shoulders to hands)
    for x in [-0.28, 0.28]:
        add_limb((x, 0, 1.45), 0.12, "Shoulder")
        add_limb((x*1.4, 0, 1.2), 0.09, "Elbow")
        add_limb((x*1.5, 0, 0.9), 0.07, "Hand")

    # LEGS
    for x in [-0.12, 0.12]:
        add_limb((x, 0, 0.7), 0.14, "Thigh")
        add_limb((x, 0, 0.4), 0.1, "Knee")
        add_limb((x, 0, 0.1), 0.08, "Ankle")
        add_limb((x, 0.1, 0.02), 0.1, "Foot")

    # HEAD (Realistic Proportions)
    add_limb((0, 0, 1.7), 0.15, "Skull")
    add_limb((0, 0.05, 1.62), 0.12, "Jaw")

    # Convert Metaball to Mesh for materials
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.object.convert(target='MESH')
    body_mesh = bpy.context.active_object
    
    # Apply Realistic Skin Material
    mat_skin = bpy.data.materials.new(name="RealisticSkin")
    mat_skin.use_nodes = True
    nodes = mat_skin.node_tree.nodes
    bsdf = nodes["Principled BSDF"]
    bsdf.inputs["Base Color"].default_value = (0.78, 0.55, 0.45, 1.0)
    # Using name instead of index for compatibility
    if "Subsurface Weight" in bsdf.inputs:
        bsdf.inputs["Subsurface Weight"].default_value = 0.1
    body_mesh.data.materials.append(mat_skin)

    # 2. Add Hair (Proper cap, not a trashcan)
    hair_col = bpy.data.collections.new("Hair")
    bpy.context.scene.collection.children.link(hair_col)
    
    bpy.ops.mesh.primitive_uv_sphere_add(radius=0.16, location=(0, -0.02, 1.75))
    hair = bpy.context.active_object
    hair.scale = (1.05, 1.1, 0.8)
    hair_col.objects.link(hair)
    
    mat_hair = bpy.data.materials.new(name="DarkHair")
    mat_hair.use_nodes = True
    mat_hair.node_tree.nodes["Principled BSDF"].inputs[0].default_value = (0.1, 0.05, 0.02, 1)
    hair.data.materials.append(mat_hair)

def setup_scene():
    bpy.ops.object.camera_add()
    cam = bpy.context.active_object
    bpy.context.scene.camera = cam
    cam.data.type = 'ORTHO'
    cam.data.ortho_scale = 2.5 # Tighter zoom for 160px
    cam.location = (10, -10, 10)
    cam.rotation_euler = (math.radians(60), 0, math.radians(45))
    
    bpy.context.scene.render.resolution_x = 160
    bpy.context.scene.render.resolution_y = 160
    bpy.context.scene.render.film_transparent = True
    
    # Three-Point Lighting for "Premium" look
    # Key light
    bpy.ops.object.light_add(type='SUN', location=(5, -5, 10))
    sun = bpy.context.active_object
    sun.data.energy = 5
    sun.rotation_euler = (math.radians(45), math.radians(15), math.radians(45))
    
    # Rim light for silhouette
    bpy.ops.object.light_add(type='SUN', location=(-5, 5, 10))
    rim = bpy.context.active_object
    rim.data.energy = 3
    rim.rotation_euler = (math.radians(-45), 0, math.radians(225))

def render_sequence():
    # Rotate character 8 times
    directions = ["n", "ne", "e", "se", "s", "sw", "w", "nw"]
    layers = [
        ("Body", "body"),
        ("Hair", "hair_brown")
    ]
    
    # Container for rotation
    bpy.ops.object.select_all(action='SELECT')
    bpy.ops.object.empty_add(type='PLAIN_AXES', location=(0,0,0))
    handle = bpy.context.active_object
    
    for obj in bpy.data.objects:
        if obj.type == 'MESH':
            obj.parent = handle

    for col in bpy.data.collections:
        col.hide_render = True
        
    for col_name, folder in layers:
        bpy.data.collections[col_name].hide_render = False
        for i, dir_name in enumerate(directions):
            handle.rotation_euler.z = math.radians(i * 45)
            path = os.path.join(BASE_DIR, folder, f"static_{dir_name}.png")
            bpy.context.scene.render.filepath = path
            bpy.ops.render.render(write_still=True)
        bpy.data.collections[col_name].hide_render = True

cleanup()
create_realistic_human()
setup_scene()
render_sequence()

bpy.ops.wm.save_as_mainfile(filepath="/Users/diegoj/repos/oinakos/tools/assets_factory/realistic_human.blend")
