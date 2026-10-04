// Package palette holds 256-entry colour tables in the style of Fractint
// .map files. Index 0 is the "inside" colour and is never cycled; indices
// 1..255 form the ring that colours escaped points.
package palette

import (
	"image/color"
	"math"
)

// Size is the number of entries in a palette.
const Size = 256

// Ring is the number of cycled entries (all but the inside colour).
const Ring = Size - 1

// Palette is a fixed table of 256 colours.
type Palette struct {
	Name   string
	Colors [Size]color.RGBA
}

// Stop is a colour at a position in [0, 1] along a gradient.
type Stop struct {
	Pos float64
	C   color.RGBA
}

// Gradient builds a palette whose ring interpolates linearly between stops.
// Stops must be sorted by Pos with the first at 0 and the last at 1. Index 0
// is set to black.
func Gradient(name string, stops ...Stop) *Palette {
	p := &Palette{Name: name}
	p.Colors[0] = color.RGBA{0, 0, 0, 255}
	if len(stops) == 0 {
		return p
	}
	for i := 1; i < Size; i++ {
		t := float64(i-1) / float64(Ring-1)
		p.Colors[i] = sample(stops, t)
	}
	return p
}

func sample(stops []Stop, t float64) color.RGBA {
	if t <= stops[0].Pos {
		return stops[0].C
	}
	for k := 1; k < len(stops); k++ {
		if t <= stops[k].Pos {
			a, b := stops[k-1], stops[k]
			span := b.Pos - a.Pos
			f := 0.0
			if span > 0 {
				f = (t - a.Pos) / span
			}
			return lerp(a.C, b.C, f)
		}
	}
	return stops[len(stops)-1].C
}

func lerp(a, b color.RGBA, f float64) color.RGBA {
	return color.RGBA{
		R: uint8(math.Round(float64(a.R) + (float64(b.R)-float64(a.R))*f)),
		G: uint8(math.Round(float64(a.G) + (float64(b.G)-float64(a.G))*f)),
		B: uint8(math.Round(float64(a.B) + (float64(b.B)-float64(a.B))*f)),
		A: 255,
	}
}

// HSV converts hue in degrees and s, v in [0, 1] to an opaque colour.
func HSV(h, s, v float64) color.RGBA {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return color.RGBA{
		R: uint8(math.Round((r + m) * 255)),
		G: uint8(math.Round((g + m) * 255)),
		B: uint8(math.Round((b + m) * 255)),
		A: 255,
	}
}
