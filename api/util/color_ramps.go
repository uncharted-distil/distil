//
//   Copyright © 2019 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package util

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/lucasb-eyer/go-colorful"
	log "github.com/unchartedsoftware/plog"
)

// BlendMode indicates the blend mode to use when building the color ramp
type BlendMode int

// RampEntry defines a colour and its location in a color ramp.
type RampEntry struct {
	ColourPoint float64
	Colour      color.RGBA
}

const (
	// RGB color space blend mode
	RGB BlendMode = iota
	// HSV color space blend mode
	HSV
	// HCL color space blend mode
	HCL
	// Lab color space blend mode
	Lab
	// None indicates no blend mode
	None
)

var (
	// RedYellowGreenRamp defines an evenly spaced ramp suitable for visualizing vegetation
	RedYellowGreenRamp = []uint8{}

	// BlueYellowBrownRamp defines an evenly spaced ramp suitable for visualizing moisture
	BlueYellowBrownRamp = []uint8{}

	// ViridisColorRamp color scale
	ViridisColorRamp = []RampEntry{}
	// MagmaColorRamp color scale
	MagmaColorRamp = []RampEntry{}
	// PlasmaColorRamp color scale
	PlasmaColorRamp = []RampEntry{}
	// InfernoColorRamp color scale
	InfernoColorRamp = []RampEntry{}
	// TurboColorRamp color scale
	TurboColorRamp = []RampEntry{}
)

func init() {
	// populate color ramps

	//RedYellowGreenRamp creates a ramp suitable for visualizing vegetation
	RedYellowGreenRamp = GenerateRamp([]RampEntry{
		{0.0, color.RGBA{162, 13, 42, 255}},
		{0.5, color.RGBA{249, 246, 179, 255}},
		{1.0, color.RGBA{16, 103, 57, 255}},
	}, 255, Lab)

	// BlueYellowBrownRamp generates a ramp sutiable for visualizing water and moisture
	BlueYellowBrownRamp = GenerateRamp([]RampEntry{
		{0.0, color.RGBA{179, 114, 59, 255}},
		{0.333, color.RGBA{243, 238, 63, 255}},
		{0.666, color.RGBA{42, 198, 223, 255}},
		{1.0, color.RGBA{5, 29, 148, 255}},
	}, 255, Lab)
	ViridisColorRamp = []RampEntry{
		{0.0, color.RGBA{68, 1, 84, 255}},
		{0.25, color.RGBA{59, 82, 139, 255}},
		{0.50, color.RGBA{33, 145, 140, 255}},
		{0.75, color.RGBA{94, 201, 98, 255}},
		{1.0, color.RGBA{253, 231, 37, 255}}}
	MagmaColorRamp = []RampEntry{
		{0.0, color.RGBA{0, 0, 4, 255}},
		{0.25, color.RGBA{81, 18, 124, 255}},
		{0.50, color.RGBA{183, 55, 121, 255}},
		{0.75, color.RGBA{252, 137, 97, 255}},
		{1.0, color.RGBA{252, 253, 191, 255}}}
	PlasmaColorRamp = []RampEntry{
		{0.0, color.RGBA{13, 8, 135, 255}},
		{0.25, color.RGBA{126, 3, 168, 255}},
		{0.50, color.RGBA{204, 71, 120, 255}},
		{0.75, color.RGBA{248, 149, 64, 255}},
		{1.0, color.RGBA{240, 249, 33, 255}}}
	InfernoColorRamp = []RampEntry{
		{0.0, color.RGBA{0, 0, 4, 255}},
		{0.25, color.RGBA{87, 16, 110, 255}},
		{0.50, color.RGBA{188, 55, 84, 255}},
		{0.75, color.RGBA{249, 142, 9, 255}},
		{1.0, color.RGBA{252, 255, 164, 255}}}
	TurboColorRamp = []RampEntry{
		{0.0, color.RGBA{35, 23, 27, 255}},
		{0.25, color.RGBA{38, 188, 225, 255}},
		{0.50, color.RGBA{149, 251, 81, 255}},
		{0.75, color.RGBA{255, 130, 29, 255}},
		{1.0, color.RGBA{144, 12, 0, 255}}}
}

// GenerateRamp creaets a a color ramp stored as a flat array of byte values.
func GenerateRamp(colors []RampEntry, steps int, blendMode BlendMode) []uint8 {
	if len(colors) == 0 || steps == 0 {
		log.Warn("no ramp entry info supplied")
		return []uint8{0, 0, 0}
	}

	ramp := []colorful.Color{}
	for i := 0; i < len(colors)-1; i++ {
		// get color point position
		pos0 := colors[i].ColourPoint
		pos1 := colors[i+1].ColourPoint

		if pos0 >= pos1 {
			log.Warn("ramp entries do not strictly increase")
			return []uint8{0, 0, 0}
		}

		// create interpolatable color at the color point
		rgbColor := colors[i].Colour
		color0 := colorful.Color{
			R: float64(rgbColor.R) / 255,
			G: float64(rgbColor.G) / 255,
			B: float64(rgbColor.B) / 255,
		}
		ramp = append(ramp, color0)

		// interpolate to the next point (without including it in the ramp)
		rgbColor = colors[i+1].Colour
		color1 := colorful.Color{
			R: float64(rgbColor.R) / 255,
			G: float64(rgbColor.G) / 255,
			B: float64(rgbColor.B) / 255,
		}

		intervalSteps := int(math.Floor((pos1 - pos0) * float64(steps-2)))
		for s := 0; s < intervalSteps; s++ {
			t := float64(s+1) / float64(intervalSteps+1)
			switch blendMode {
			case RGB:
				ramp = append(ramp, color0.BlendRgb(color1, t))
			case HSV:
				ramp = append(ramp, color0.BlendHsv(color1, t))
			case HCL:
				ramp = append(ramp, color0.BlendHcl(color1, t))
			case Lab:
				ramp = append(ramp, color0.BlendLab(color1, t))
			}
		}
	}
	// add the final color point
	rgbColor := colors[len(colors)-1].Colour
	lastColor := colorful.Color{
		R: float64(rgbColor.R) / 255,
		G: float64(rgbColor.G) / 255,
		B: float64(rgbColor.B) / 255}
	ramp = append(ramp, lastColor)

	// convert result into uint rgb
	result := make([]uint8, len(ramp)*3)
	for i, color := range ramp {
		r, g, b := color.RGB255()
		result[i*3] = r
		result[i*3+1] = g
		result[i*3+2] = b
	}
	return result
}

// GetColor is a color scale function for normalized values
func GetColor(normalizedVal float64, ramp []RampEntry) *color.RGBA {
	for i := 0; i < len(ramp)-1; i++ {
		c1 := ramp[i]
		c2 := ramp[i+1]
		if c1.ColourPoint <= normalizedVal && normalizedVal <= c2.ColourPoint {
			// We are in between c1 and c2. Go blend them!
			delta := (normalizedVal - c1.ColourPoint) / (c2.ColourPoint - c1.ColourPoint)
			col1 := colorful.Color{
				R: float64(c1.Colour.R) / 255,
				G: float64(c1.Colour.G) / 255,
				B: float64(c1.Colour.B) / 255,
			}
			col2 := colorful.Color{
				R: float64(c2.Colour.R) / 255,
				G: float64(c2.Colour.G) / 255,
				B: float64(c2.Colour.B) / 255,
			}
			result := col1.BlendHcl(col2, delta).Clamped()
			return &color.RGBA{uint8(result.R * 255), uint8(result.G * 255), uint8(result.B * 255), 255}
		}
	}

	// Nothing found? Means we're at (or past) the last gradient keypoint.
	return &ramp[len(ramp)-1].Colour
}

// ViridisColorScale returns a functions used to return a color from the viridis color scale given a normalized value
func ViridisColorScale(normalizedVal float64) *color.RGBA {
	return GetColor(normalizedVal, ViridisColorRamp)
}

// MagmaColorScale returns a functions used to return a color from the magma color scale given a normalized value
func MagmaColorScale(normalizedVal float64) *color.RGBA {
	return GetColor(normalizedVal, MagmaColorRamp)
}

// PlasmaColorScale returns a functions used to return a color from the plasma color scale given a normalized value
func PlasmaColorScale(normalizedVal float64) *color.RGBA {
	return GetColor(normalizedVal, PlasmaColorRamp)
}

// InfernoColorScale returns a functions used to return a color from the inferno color scale given a normalized value
func InfernoColorScale(normalizedVal float64) *color.RGBA {
	return GetColor(normalizedVal, InfernoColorRamp)
}

// TurboColorScale returns a functions used to return a color from the inferno color scale given a normalized value
func TurboColorScale(normalizedVal float64) *color.RGBA {
	return GetColor(normalizedVal, TurboColorRamp)
}

// GetColorScale returns the color scale function based the supplied name. if name is incorrect defaults to viridis function
func GetColorScale(colorScaleName string) func(float64) *color.RGBA {
	switch colorScaleName {
	case "viridis":
		return ViridisColorScale
	case "magma":
		return MagmaColorScale
	case "plasma":
		return PlasmaColorScale
	case "inferno":
		return InfernoColorScale
	case "turbo":
		return TurboColorScale
	default:
		return ViridisColorScale
	}
}

// RampToImage converts a color ramp to an image for debugging purposes
func RampToImage(height int, ramp []uint8) *image.RGBA {
	blocks := len(ramp) / 3
	blockw := 10
	img := image.NewRGBA(image.Rect(0, 0, blocks*blockw, height))
	colorIdx := 0
	for i := 0; i < blocks; i++ {
		draw.Draw(img, image.Rect(i*blockw, 0, (i+1)*blockw, height), &image.Uniform{color.RGBA{ramp[colorIdx],
			ramp[colorIdx+1], ramp[colorIdx+2], 255}}, image.Point{}, draw.Src)
		colorIdx += 3
	}
	return img
}
