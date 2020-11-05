package util

import "math"

// Options for ConvertS2ToRgb
type Options struct {
	gain  float64
	gamma float64
	gainL float64
}

// ConvertS2ToRgb bands: [b02, b03, b04], Options gain amount, gamma correction amount, gainL light gain
// default options: gain=2.5, gamma=2.2, gainL=1.0
func ConvertS2ToRgb(bands [3]float64, options ...Options) [3]float64 {
	t := [3][3]float64{{0.268, 0.362, 0.371},
		{0.240, 0.587, 0.174},
		{1.463, -0.427, -0.043}} // magic matrix
	if len(options) != 0 { // if options
		opt := options[0]
		return s2ToRGB(getSolarIrr(bands[2], bands[1], bands[0]), t, opt.gain, opt.gamma, opt.gainL)
	}
	// defaults found here:
	// https://github.com/sentinel-hub/custom-scripts/blob/e16f0d4f52fc2f9aaf612e582865d17b0b5c3457/sentinel-2/natural_color/script.js#L92
	gain := 2.5  // looks like color scalar
	gamma := 2.2 // increases lightness by exponential
	gainL := 1.0 // increases lightness by factor
	return s2ToRGB(getSolarIrr(bands[2], bands[1], bands[0]), t, gain, gamma, gainL)
}
func s2ToRGB(rad [3]float64, T [3][3]float64, gain float64, gamma float64, gL float64) [3]float64 {
	var XYZ = s2ToXYZ(rad, T, gain)
	var Lab = xyzToLab(XYZ)
	var L = math.Pow(gL*Lab[0], gamma)
	return labToRGB([3]float64{L, Lab[1], Lab[2]})
}

// vector * scalar
func dotVS(vec [3]float64, scalar float64) [3]float64 {
	result := [3]float64{0, 0, 0}
	for i, v := range vec {
		result[i] = v * scalar
	}
	return result
}

// dot product of two vectors
func dotVV(vec1 [3]float64, vec2 [3]float64) float64 {
	result := 0.0
	for i, v := range vec1 {
		result += v * vec2[i]
	}
	return result
}

// matrix . vector
func dotMV(mat [3][3]float64, vec [3]float64) [3]float64 {
	result := [3]float64{0, 0, 0}
	for i, v := range mat {
		result[i] = dotVV(v, vec)
	}
	return result
}

func s2ToXYZ(rad [3]float64, T [3][3]float64, gain float64) [3]float64 {
	return dotVS(dotMV(T, rad), gain)
}

func getSolarIrr(b02 float64, b03 float64, b04 float64) [3]float64 { // possible solar Irradiance adjustment
	return [3]float64{b02, 0.939 * b03, 0.779 * b04} // some sort of magic value to convet to irr
}
func labF(val float64) float64 {
	if val > 0.00885645 {
		return math.Pow(val, 1.0/3.0)
	}
	return 0.137931 + 7.787*val
}
func invLabF(t float64) float64 {
	if t > 0.2069 {
		return t * t * t
	}
	return 0.12842 * (t - 0.137931)
}
func xyzToLab(XYZ [3]float64) [3]float64 {
	var lfY = labF(XYZ[1])
	return [3]float64{(116.0*lfY - 16) / 100,
		5 * (labF(XYZ[0]) - lfY),
		2 * (lfY - labF(XYZ[2]))}
}

func labToRGB(Lab [3]float64) [3]float64 {
	return xyzToRGB(labToXYZ(Lab))
}

func labToXYZ(Lab [3]float64) [3]float64 {
	var YL = (100*Lab[0] + 16) / 116
	return [3]float64{invLabF(YL + Lab[1]/5.0),
		invLabF(YL),
		invLabF(YL - Lab[2]/2.0)}
}
func adj(C float64) float64 {
	if C <= 0.0031308 {
		return 12.92 * C
	}
	return 1.055*math.Pow(C, 1.0/2.4) - 0.055
}
func xyzToRGBlin(xyz [3]float64) [3]float64 {
	return dotMV([3][3]float64{{3.2404542, -1.5371385, -0.4985314}, {-0.9692660, 1.8760108, 0.0415560}, {0.0556434, -0.2040259, 1.0572252}}, xyz)
}

func xyzToRGB(xyz [3]float64) [3]float64 {
	sRGB := xyzToRGBlin(xyz)
	for i, v := range sRGB {
		sRGB[i] = math.Max(math.Min(adj(v), 1.0), 0)
	}
	return sRGB
}
