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

	// Primary (Magenta #FF00FF)
	if abs(imgColor.r - 1.0) < 0.01 && abs(imgColor.g - 0.0) < 0.01 && abs(imgColor.b - 1.0) < 0.01 {
		return PrimaryColor * imgColor.a
	}
	// Secondary (Yellow #FFFF00)
	if abs(imgColor.r - 1.0) < 0.01 && abs(imgColor.g - 1.0) < 0.01 && abs(imgColor.b - 0.0) < 0.01 {
		return SecondaryColor * imgColor.a
	}

	return imgColor
}
`)
