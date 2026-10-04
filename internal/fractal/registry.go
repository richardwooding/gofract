package fractal

import "fmt"

// kernel iterates from z0 = (zr, zi) with parameter c = (cr, ci).
type kernel func(zr, zi, cr, ci float64, maxIter int) float32

// family is one formula with its default views.
type family struct {
	name     string
	k        kernel
	hasJulia bool
	// Mandelbrot-mode view and a sensible Julia constant.
	cx, cy, scale float64
	jre, jim      float64
	// Julia-mode view.
	jcx, jcy, jscale float64
	// shortcut reports points known to be inside, in Mandelbrot mode.
	shortcut func(x, y float64) bool
}

func (f *family) start() Params {
	return Params{
		Type: f.name, CenterX: Dec(f.cx), CenterY: Dec(f.cy), Scale: f.scale,
		MaxIter: 256, JuliaRe: f.jre, JuliaIm: f.jim,
	}
}

var families = []*family{
	{name: "mandelbrot", k: mandelbrot, hasJulia: true,
		cx: -0.5, cy: 0, scale: 3.5, jre: -0.8, jim: 0.156, jcx: 0, jcy: 0, jscale: 4,
		shortcut: inMainBulbs},
	{name: "burningship", k: burningShip, hasJulia: true,
		cx: -0.45, cy: 0.55, scale: 3.6, jre: -0.6, jim: 0.9, jcx: 0, jcy: 0, jscale: 4},
	{name: "tricorn", k: tricorn, hasJulia: true,
		cx: -0.3, cy: 0, scale: 4, jre: 0.3, jim: 0.5, jcx: 0, jcy: 0, jscale: 4},
	{name: "multibrot3", k: multibrot3, hasJulia: true,
		cx: 0, cy: 0, scale: 3.2, jre: 0.4, jim: 0.3, jcx: 0, jcy: 0, jscale: 3.5},
	{name: "newton", k: newton, hasJulia: false,
		cx: 0, cy: 0, scale: 4},
}

func lookup(name string) (*family, error) {
	for _, f := range families {
		if f.name == name {
			return f, nil
		}
	}
	return nil, fmt.Errorf("fractal: unknown type %q", name)
}

// Names lists the registered families in menu order.
func Names() []string {
	out := make([]string, len(families))
	for i, f := range families {
		out[i] = f.name
	}
	return out
}

// HasJulia reports whether the family has a Julia-mode variant.
func HasJulia(name string) bool {
	f, err := lookup(name)
	return err == nil && f.hasJulia
}

// Start returns the default view of the named family.
func Start(name string) (Params, error) {
	f, err := lookup(name)
	if err != nil {
		return Params{}, err
	}
	return f.start(), nil
}

// New builds the Fractal selected by p.Type and p.Julia.
func New(p Params) (Fractal, error) {
	// Accept the v1 sidecar spelling where Julia was its own type.
	if p.Type == "julia" {
		p.Type, p.Julia = "mandelbrot", true
	}
	if p.Type == "" {
		p.Type = "mandelbrot"
	}
	f, err := lookup(p.Type)
	if err != nil {
		return nil, err
	}
	cx, cy := p.CenterF()
	if p.Julia {
		if !f.hasJulia {
			return nil, fmt.Errorf("fractal: %s has no Julia variant", f.name)
		}
		return &juliaSet{name: f.name, k: f.k, cx: cx, cy: cy, cr: p.JuliaRe, ci: p.JuliaIm}, nil
	}
	if p.Deep() && f.name == "mandelbrot" {
		return newPerturb(p), nil
	}
	return &mandelSet{name: f.name, k: f.k, cx: cx, cy: cy, shortcut: f.shortcut}, nil
}

// HasDeep reports whether the family can render below float64 precision.
func HasDeep(p Params) bool { return p.Type == "mandelbrot" && !p.Julia }

type mandelSet struct {
	name     string
	k        kernel
	cx, cy   float64
	shortcut func(x, y float64) bool
}

func (m *mandelSet) Name() string { return m.name }

func (m *mandelSet) Iterate(dx, dy float64, maxIter int) float32 {
	x, y := m.cx+dx, m.cy+dy
	if m.shortcut != nil && m.shortcut(x, y) {
		return Inside
	}
	return m.k(0, 0, x, y, maxIter)
}

type juliaSet struct {
	name   string
	k      kernel
	cx, cy float64
	cr, ci float64
}

func (j *juliaSet) Name() string { return j.name }

func (j *juliaSet) Iterate(dx, dy float64, maxIter int) float32 {
	return j.k(j.cx+dx, j.cy+dy, j.cr, j.ci, maxIter)
}

// ToJulia switches to Julia mode with the given plane point as the constant,
// framed at the family's default Julia view. This is Fractint's Space-bar
// behaviour. Families without a Julia variant are returned unchanged.
func (p Params) ToJulia(cx, cy float64) Params {
	f, err := lookup(p.Type)
	if err != nil || !f.hasJulia {
		return p
	}
	p.Julia = true
	p.JuliaRe, p.JuliaIm = cx, cy
	p.CenterX, p.CenterY, p.Scale = Dec(f.jcx), Dec(f.jcy), f.jscale
	return p
}

// ToMandelbrot leaves Julia mode, centred on the Julia constant so the user
// lands where they left.
func (p Params) ToMandelbrot() Params {
	f, err := lookup(p.Type)
	if err != nil {
		return p
	}
	p.Julia = false
	p.CenterX, p.CenterY, p.Scale = Dec(p.JuliaRe), Dec(p.JuliaIm), f.scale
	return p
}

// WithType switches family, keeping max iterations but resetting the view
// to the family's default.
func (p Params) WithType(name string) (Params, error) {
	f, err := lookup(name)
	if err != nil {
		return p, err
	}
	q := f.start()
	q.MaxIter = p.MaxIter
	return q, nil
}

// TypeIndex returns the family's position in Names(), or -1 if unknown. The
// GPU shader selects its kernel by this index.
func TypeIndex(name string) int {
	for i, f := range families {
		if f.name == name {
			return i
		}
	}
	return -1
}
