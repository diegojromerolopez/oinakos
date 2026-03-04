package game

var paletteSwapShaderSource = []byte(`
//kage:unit pixels

package main

var PrimaryColor vec4
var SecondaryColor vec4

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	imgColor := imageSrc0At(srcPos)
	if imgColor.a == 0.0 {
		return imgColor
	}

	// Normalize color for mask checking (ignore alpha)
	// Add a small epsilon to avoids division by zero or large errors
	rawRGB := imgColor.rgb / (imgColor.a + 0.00001)

	// Primary (Magenta #FF00FF)
	distMagenta := distance(rawRGB, vec3(1.0, 0.0, 1.0))
	// User's custom reported color: rgb(60, 111, 227/255)
	distCustom1 := distance(rawRGB, vec3(60.0/255.0, 111.0/255.0, 227.0/255.0))
	distCustom2 := distance(rawRGB, vec3(60.0/255.0, 111.0/255.0, 1.0)) // if 327 meant maxed out blue
	distCustom3 := distance(rawRGB, vec3(60.0/255.0, 11.0/255.0, 227.0/255.0)) // if 111 meant 11

	if distMagenta < 0.45 || distCustom1 < 0.45 || distCustom2 < 0.45 || distCustom3 < 0.45 {
		return PrimaryColor * imgColor.a
	}
	
	// Secondary (Yellow #FFFF00)
	if distance(rawRGB, vec3(1.0, 1.0, 0.0)) < 0.45 {
		return SecondaryColor * imgColor.a
	}

	return imgColor
}
`)
