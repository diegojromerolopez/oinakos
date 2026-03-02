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
	rawRGB := imgColor.rgb / imgColor.a

	// Primary (Magenta #FF00FF)
	if distance(rawRGB, vec3(1.0, 0.0, 1.0)) < 0.1 {
		return PrimaryColor * imgColor.a
	}
	// Secondary (Yellow #FFFF00)
	if distance(rawRGB, vec3(1.0, 1.0, 0.0)) < 0.1 {
		return SecondaryColor * imgColor.a
	}

	return imgColor
}
`)
