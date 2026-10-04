package palette

import "image/color"

func rgb(r, g, b uint8) color.RGBA { return color.RGBA{r, g, b, 255} }

// Presets returns the built-in palettes in menu order. The first entry is
// the default.
func Presets() []*Palette {
	return []*Palette{
		classic(),
		Gradient("fire",
			Stop{0, rgb(0, 0, 0)},
			Stop{0.25, rgb(128, 0, 0)},
			Stop{0.5, rgb(255, 64, 0)},
			Stop{0.75, rgb(255, 200, 0)},
			Stop{1, rgb(255, 255, 255)},
		),
		Gradient("ocean",
			Stop{0, rgb(0, 0, 32)},
			Stop{0.3, rgb(0, 32, 160)},
			Stop{0.6, rgb(0, 192, 255)},
			Stop{0.85, rgb(220, 250, 255)},
			Stop{1, rgb(0, 0, 32)},
		),
		Gradient("grey",
			Stop{0, rgb(0, 0, 0)},
			Stop{1, rgb(255, 255, 255)},
		),
		rainbow(),
	}
}

// classic approximates Fractint's default.map: a smooth cycle through blues,
// cyans, greens, yellows and reds that wraps on itself so palette cycling is
// seamless.
func classic() *Palette {
	return Gradient("classic",
		Stop{0, rgb(0, 7, 100)},
		Stop{0.16, rgb(32, 107, 203)},
		Stop{0.42, rgb(237, 255, 255)},
		Stop{0.6425, rgb(255, 170, 0)},
		Stop{0.8575, rgb(0, 2, 0)},
		Stop{1, rgb(0, 7, 100)},
	)
}

func rainbow() *Palette {
	p := &Palette{Name: "rainbow"}
	p.Colors[0] = rgb(0, 0, 0)
	for i := 1; i < Size; i++ {
		h := 360 * float64(i-1) / float64(Ring)
		p.Colors[i] = HSV(h, 1, 1)
	}
	return p
}
