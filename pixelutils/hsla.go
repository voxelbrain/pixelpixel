package pixelutils

// This code is a modified version of github.com/bthomson/go-color

// Implements image/color.Color for HSLA values
type HSLA struct {
	H, S, L, A float64
}

func (c HSLA) RGBA() (r, g, b, a uint32) {
	if c.S == 0 {
		l := floatToUint32(c.L)
		return l, l, l, floatToUint32(c.A)
	}

	var v1, v2 float64
	if c.L < 0.5 {
		v2 = c.L * (1 + c.S)
	} else {
		v2 = (c.L + c.S) - (c.S * c.L)
	}

	v1 = 2*c.L - v2

	r = floatToUint32(hueToRGB(v1, v2, c.H+(1.0/3.0)))
	g = floatToUint32(hueToRGB(v1, v2, c.H))
	b = floatToUint32(hueToRGB(v1, v2, c.H-(1.0/3.0)))
	a = floatToUint32(c.A)
	return
}

func floatToUint32(f float64) uint32 {
	return uint32(float64(0xFFFF) * f)
}

func hueToRGB(v1, v2, h float64) float64 {
	if h < 0 {
		h += 1
	}
	if h > 1 {
		h -= 1
	}
	switch {
	case 6*h < 1:
		return (v1 + (v2-v1)*6*h)
	case 2*h < 1:
		return v2
	case 3*h < 2:
		return v1 + (v2-v1)*((2.0/3.0)-h)*6
	}
	return v1
}
