import bpy
import math
import os

# Configuration
OUTPUT_DIR = "/Users/diegoj/repos/oinakos/assets/images/obstacles/trestle_table"
os.makedirs(OUTPUT_DIR, exist_ok=True)

def cleanup():
    bpy.ops.wm.read_factory_settings(use_empty=True)

def create_table():
    # Create the top
    bpy.ops.mesh.primitive_cube_add(size=1, location=(0, 0, 0.8))
    table_top = bpy.context.active_object
    table_top.scale = (1.5, 0.8, 0.05)
    
    # Create legs (trestle style)
    for x_side in [-1.2, 1.2]:
        bpy.ops.mesh.primitive_cube_add(size=1, location=(x_side, 0, 0.4))
        leg = bpy.context.active_object
        leg.scale = (0.05, 0.6, 0.4)
        
    # Add a support beam
    bpy.ops.mesh.primitive_cube_add(size=1, location=(0, 0, 0.4))
    beam = bpy.context.active_object
    beam.scale = (1.2, 0.05, 0.05)

def setup_lighting():
    # Clear existing lights
    bpy.ops.object.select_by_type(type='LIGHT')
    bpy.ops.object.delete()
    
    # Add Sun light for that "Diablo" outdoor look
    bpy.ops.object.light_add(type='SUN', location=(5, -5, 10))
    sun = bpy.context.active_object
    sun.data.energy = 5
    sun.rotation_euler = (math.radians(45), 0, math.radians(45))

def setup_camera():
    bpy.ops.object.camera_add()
    cam = bpy.context.active_object
    bpy.context.scene.camera = cam
    
    cam.data.type = 'ORTHO'
    cam.data.ortho_scale = 5.0
    cam.location = (10, -10, 10)
    # Isometric look: 60 deg X, 45 deg Z
    cam.rotation_euler = (math.radians(60), 0, math.radians(45))
    
    bpy.context.scene.render.resolution_x = 160
    bpy.context.scene.render.resolution_y = 160
    bpy.context.scene.render.film_transparent = True

def render_8_directions():
    # We rotate the table, not the camera
    # Place all objects in a container to rotate them together
    bpy.ops.object.select_all(action='SELECT')
    
    # Create an empty to use as a rotation handle
    bpy.ops.object.empty_add(type='PLAIN_AXES', location=(0,0,0))
    parent_empty = bpy.context.active_object
    parent_empty.name = "RotationHandle"
    
    # Select all meshes and parent them to the empty
    for obj in bpy.data.objects:
        if obj.type == 'MESH':
            obj.select_set(True)
        else:
            obj.select_set(False)
    
    parent_empty.select_set(True)
    bpy.context.view_layer.objects.active = parent_empty
    bpy.ops.object.parent_set(type='OBJECT')

    directions = ["n", "ne", "e", "se", "s", "sw", "w", "nw"]
    
    for i, name in enumerate(directions):
        parent_empty.rotation_euler.z = math.radians(i * 45)
        
        filepath = os.path.join(OUTPUT_DIR, f"static_{name}.png")
        bpy.context.scene.render.filepath = filepath
        bpy.ops.render.render(write_still=True)

cleanup()
create_table()
setup_lighting()
setup_camera()
render_8_directions()

# Save the blend file so the user can open it!
bpy.ops.wm.save_as_mainfile(filepath="/Users/diegoj/repos/oinakos/tools/assets_factory/table_gen.blend")
