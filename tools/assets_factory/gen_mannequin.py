import bpy
import math
import os

# Configuration
BASE_DIR = "/Users/diegoj/repos/oinakos/assets/images/characters/human_male"
os.makedirs(os.path.join(BASE_DIR, "body"), exist_ok=True)

def cleanup():
    bpy.ops.wm.read_factory_settings(use_empty=True)

def create_block_mannequin():
    collection = bpy.data.collections.new("Body")
    bpy.context.scene.collection.children.link(collection)
    
    # 1. TORSO (The "box" method)
    # Pelvis
    bpy.ops.mesh.primitive_cube_add(size=1, location=(0, 0, 0.95))
    hips = bpy.context.active_object
    hips.scale = (0.2, 0.12, 0.15)
    
    # Chest (V-taper)
    bpy.ops.mesh.primitive_cube_add(size=1, location=(0, 0, 1.35))
    chest = bpy.context.active_object
    chest.scale = (0.25, 0.14, 0.22)
    
    # 2. HEAD (Ovoid)
    bpy.ops.mesh.primitive_uv_sphere_add(radius=0.14, location=(0, 0, 1.7))
    head = bpy.context.active_object
    head.scale = (0.85, 1, 1.1) # Human skull shape

    # 3. LEGS (Articulated)
    for x in [-0.11, 0.11]:
        # Upper Leg
        bpy.ops.mesh.primitive_cylinder_add(radius=0.08, depth=0.45, location=(x, 0, 0.72))
        u_leg = bpy.context.active_object
        u_leg.rotation_euler.y = math.radians(2 if x > 0 else -2)
        
        # Lower Leg
        bpy.ops.mesh.primitive_cylinder_add(radius=0.06, depth=0.45, location=(x, 0, 0.28))
        l_leg = bpy.context.active_object
        
        # Foot
        bpy.ops.mesh.primitive_cube_add(size=1, location=(x, 0.08, 0.05))
        foot = bpy.context.active_object
        foot.scale = (0.07, 0.12, 0.04)

    # 4. ARMS
    for x in [-0.3, 0.3]:
        # Shoulder "Ball"
        bpy.ops.mesh.primitive_uv_sphere_add(radius=0.08, location=(x, 0, 1.5))
        
        # Upper Arm
        bpy.ops.mesh.primitive_cylinder_add(radius=0.06, depth=0.35, location=(x, 0, 1.32))
        
        # Lower Arm
        bpy.ops.mesh.primitive_cylinder_add(radius=0.05, depth=0.35, location=(x, 0, 1.0))
        
        # Hand (Block)
        bpy.ops.mesh.primitive_cube_add(size=1, location=(x, 0, 0.8))
        hand = bpy.context.active_object
        hand.scale = (0.05, 0.05, 0.06)

    # 5. Global Material (Clean Clay)
    mat = bpy.data.materials.new(name="MannequinClay")
    mat.use_nodes = True
    mat.node_tree.nodes["Principled BSDF"].inputs["Base Color"].default_value = (0.6, 0.6, 0.6, 1)
    mat.node_tree.nodes["Principled BSDF"].inputs["Roughness"].default_value = 0.4

    for obj in bpy.data.objects:
        if obj.type == 'MESH':
            obj.data.materials.append(mat)
            collection.objects.link(obj)
            # Remove from default collection
            if obj.name in bpy.context.scene.collection.objects:
                bpy.context.scene.collection.objects.unlink(obj)

def setup_camera():
    bpy.ops.object.camera_add()
    cam = bpy.context.active_object
    bpy.context.scene.camera = cam
    cam.data.type = 'ORTHO'
    cam.data.ortho_scale = 2.8
    cam.location = (10, -10, 10)
    cam.rotation_euler = (math.radians(60), 0, math.radians(45))
    
    bpy.context.scene.render.resolution_x = 160
    bpy.context.scene.render.resolution_y = 160
    bpy.context.scene.render.film_transparent = True
    
    # Dramatic Top-Down Lighting
    bpy.ops.object.light_add(type='SUN', location=(5, -5, 10))
    sun = bpy.context.active_object
    sun.data.energy = 4
    sun.rotation_euler = (math.radians(45), math.radians(20), math.radians(45))

def render_8_directions():
    directions = ["n", "ne", "e", "se", "s", "sw", "w", "nw"]
    
    # Select all meshes
    bpy.ops.object.select_all(action='SELECT')
    bpy.ops.object.empty_add(type='PLAIN_AXES', location=(0,0,0))
    handle = bpy.context.active_object
    
    for obj in bpy.data.objects:
        if obj.type == 'MESH':
            obj.parent = handle

    for i, dir_name in enumerate(directions):
        handle.rotation_euler.z = math.radians(i * 45)
        path = os.path.join(BASE_DIR, "body", f"static_{dir_name}.png")
        bpy.context.scene.render.filepath = path
        bpy.ops.render.render(write_still=True)

cleanup()
create_block_mannequin()
setup_camera()
render_8_directions()

bpy.ops.wm.save_as_mainfile(filepath="/Users/diegoj/repos/oinakos/tools/assets_factory/mannequin_human.blend")
