package fractal

import "fmt"

// Names lists the registered fractal types in menu order.
func Names() []string {
	return []string{"mandelbrot", "julia"}
}

// New builds the Fractal selected by p.Type.
func New(p Params) (Fractal, error) {
	switch p.Type {
	case "mandelbrot", "":
		return Mandelbrot{}, nil
	case "julia":
		return Julia{C: complex(p.JuliaRe, p.JuliaIm)}, nil
	}
	return nil, fmt.Errorf("fractal: unknown type %q", p.Type)
}

// ToJulia returns params for the Julia set whose constant is the given plane
// point, framed at the origin. This is Fractint's Space-bar behaviour.
func (p Params) ToJulia(cx, cy float64) Params {
	p.Type = "julia"
	p.JuliaRe = cx
	p.JuliaIm = cy
	p.CenterX = 0
	p.CenterY = 0
	p.Scale = 4
	return p
}

// ToMandelbrot returns params back on the Mandelbrot set, centred on the
// Julia constant so the user lands where they left.
func (p Params) ToMandelbrot() Params {
	p.Type = "mandelbrot"
	p.CenterX = p.JuliaRe
	p.CenterY = p.JuliaIm
	p.Scale = 3.5
	return p
}
