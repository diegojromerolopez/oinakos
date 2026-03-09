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

def create_human_base():
    # 1. Body Collection
    body_col = bpy.data.collections.new("Body")
    bpy.context.scene.collection.children.link(body_col)
    
    # Torso
    bpy.ops.mesh.primitive_cube_add(size=1, location=(0, 0, 1.2))
    torso = bpy.context.active_object
    torso.name = "Torso"
    torso.scale = (0.25, 0.15, 0.4)
    body_col.objects.link(torso)
    
    # Head
    bpy.ops.mesh.primitive_uv_sphere_add(radius=0.15, location=(0, 0, 1.8))
    head = bpy.context.active_object
    head.name = "Head"
    body_col.objects.link(head)
    
    # Legs
    for x in [-0.1, 0.1]:
        bpy.ops.mesh.primitive_cylinder_add(radius=0.06, depth=0.8, location=(x, 0, 0.4))
        leg = bpy.context.active_object
        body_col.objects.link(leg)
        
    # 2. Clothing Collection
    clothing_col = bpy.data.collections.new("Tunic")
    bpy.context.scene.collection.children.link(clothing_col)
    
    # Simple Tunic (slightly larger than torso)
    bpy.ops.mesh.primitive_cube_add(size=1, location=(0, 0, 1.22))
    tunic = bpy.context.active_object
    tunic.name = "Tunic_Blue"
    tunic.scale = (0.27, 0.17, 0.38)
    clothing_col.objects.link(tunic)

    # 3. Hair Collection
    hair_col = bpy.data.collections.new("Hair")
    bpy.context.scene.collection.children.link(hair_col)
    
    # Simple Hair Cap
    bpy.ops.mesh.primitive_uv_sphere_add(radius=0.16, location=(0, 0, 1.85))
    hair = bpy.context.active_object
    hair.name = "Hair_Brown"
    hair.scale = (1, 1, 0.6)
    hair_col.objects.link(hair)

    # Setup Materials
    mat_skin = bpy.data.materials.new(name="Skin")
    mat_skin.use_nodes = True
    mat_skin.node_tree.nodes["Principled BSDF"].inputs[0].default_value = (0.8, 0.6, 0.5, 1) # Peach
    
    mat_hair = bpy.data.materials.new(name="Hair")
    mat_hair.use_nodes = True
    mat_hair.node_tree.nodes["Principled BSDF"].inputs[0].default_value = (0.2, 0.1, 0.05, 1) # Dark Brown
    
    mat_tunic = bpy.data.materials.new(name="Tunic")
    mat_tunic.use_nodes = True
    mat_tunic.node_tree.nodes["Principled BSDF"].inputs[0].default_value = (0.1, 0.2, 0.5, 1) # Blue
    
    torso.data.materials.append(mat_skin)
    head.data.materials.append(mat_skin)
    hair.data.materials.append(mat_hair)
    tunic.data.materials.append(mat_tunic)
    for obj in body_col.objects:
        if obj.name.startswith("Cylinder"): obj.data.materials.append(mat_skin)

def setup_scene():
    bpy.ops.object.camera_add()
    cam = bpy.context.active_object
    bpy.context.scene.camera = cam
    cam.data.type = 'ORTHO'
    cam.data.ortho_scale = 3.0
    cam.location = (10, -10, 10)
    cam.rotation_euler = (math.radians(60), 0, math.radians(45))
    
    bpy.context.scene.render.resolution_x = 160
    bpy.context.scene.render.resolution_y = 160
    bpy.context.scene.render.film_transparent = True
    
    bpy.ops.object.light_add(type='SUN', location=(5, -5, 10))
    bpy.context.active_object.data.energy = 4

def render_sequence():
    # Create rotation handle
    bpy.ops.object.empty_add(type='PLAIN_AXES', location=(0,0,0))
    handle = bpy.context.active_object
    
    # Parent ALL meshes to handle
    for obj in bpy.data.objects:
        if obj.type == 'MESH':
            obj.parent = handle

    directions = ["n", "ne", "e", "se", "s", "sw", "w", "nw"]
    layers = [
        ("Body", "body"),
        ("Hair", "hair_brown"),
        ("Tunic", "tunic_blue")
    ]
    
    # Hide everything first
    for col in bpy.data.collections:
        col.hide_render = True
        
    for col_name, folder in layers:
        # Show only this layer
        bpy.data.collections[col_name].hide_render = False
        
        for i, dir_name in enumerate(directions):
            handle.rotation_euler.z = math.radians(i * 45)
            
            path = os.path.join(BASE_DIR, folder, f"static_{dir_name}.png")
            bpy.context.scene.render.filepath = path
            bpy.ops.render.render(write_still=True)
            
        # Hide it again for next loop
        bpy.data.collections[col_name].hide_render = True

cleanup()
create_human_base()
setup_scene()
render_sequence()

bpy.ops.wm.save_as_mainfile(filepath="/Users/diegoj/repos/oinakos/tools/assets_factory/human_male_rig.blend")
