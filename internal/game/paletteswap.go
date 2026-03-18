package game

var paletteSwapShaderSource = []byte(`
//kage:unit pixels

package main

var PrimaryColor vec4
var SecondaryColor vec4

// Trauma Flags (using floats for compatibility)
// 1.0 = true, 0.0 = false
var LeftArmLost float
var RightArmLost float
var LeftLegLost float
var RightLegLost float
var BurnedAlive float
var EyesLost float

// Status Tint for effects like Poison (green) or Blood loss (pale)
var StatusTint vec4

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	imgColor := imageSrc0At(srcPos)
	if imgColor.a == 0.0 {
		return imgColor
	}

	// Geometric Clipping (Amputations)
	// Sprite is 160x160. Normalizing coordinates is done by dividing by image size (160)
	// though kage 'pixels' unit gives us raw coords.
	
	// Legs (Bottom roughly Y > 115)
	if LeftLegLost > 0.5 && srcPos.y > 115.0 && srcPos.x < 80.0 {
		return vec4(0)
	}
	if RightLegLost > 0.5 && srcPos.y > 115.0 && srcPos.x >= 80.0 {
		return vec4(0)
	}

	// Arms (Sides roughly X < 45 or X > 115)
	// Arm clipping thresholds depend on the pose, but centering on the body (80)
	if LeftArmLost > 0.5 && srcPos.x < 55.0 && srcPos.y < 115.0 {
		return vec4(0)
	}
	if RightArmLost > 0.5 && srcPos.x > 105.0 && srcPos.y < 115.0 {
		return vec4(0)
	}

	// Eye Loss (Very rough masking around the face center top)
	// This is experimental and might need specific sprite-dependent regions
	if EyesLost >= 1.0 && srcPos.y > 40.0 && srcPos.y < 55.0 && srcPos.x > 75.0 && srcPos.x < 85.0 {
		// Just turn the eye area black if one is missing
		imgColor = vec4(0, 0, 0, imgColor.a)
	}

	// Normalize color for mask checking (ignore alpha)
	rawRGB := imgColor.rgb / (imgColor.a + 0.00001)

	// Primary Palette Swap (Magenta #FF00FF)
	distMagenta := distance(rawRGB, vec3(1.0, 0.0, 1.0))
	distCustom1 := distance(rawRGB, vec3(60.0/255.0, 111.0/255.0, 227.0/255.0))
	distCustom2 := distance(rawRGB, vec3(60.0/255.0, 111.0/255.0, 1.0))
	distCustom3 := distance(rawRGB, vec3(60.0/255.0, 11.0/255.0, 227.0/255.0))

	if distMagenta < 0.45 || distCustom1 < 0.45 || distCustom2 < 0.45 || distCustom3 < 0.45 {
		imgColor = PrimaryColor * imgColor.a
	}
	
	// Secondary Palette Swap (Yellow #FFFF00)
	if distance(rawRGB, vec3(1.0, 1.0, 0.0)) < 0.45 {
		imgColor = SecondaryColor * imgColor.a
	}

	// Burned Alive (Apply charred darkening)
	if BurnedAlive > 0.5 {
		imgColor.rgb *= 0.4 // Darken significantly
		imgColor.r += 0.1   // Add slight ember red glow
	}

	// Apply Final Tints (Poison, etc)
	if StatusTint.a > 0.0 {
		imgColor.rgb = mix(imgColor.rgb, StatusTint.rgb * imgColor.a, StatusTint.a)
	}

	return imgColor
}
`)
