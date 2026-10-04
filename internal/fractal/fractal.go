// Package fractal defines the fractal families gofract can render and the
// view parameters that select a region of the complex plane.
package fractal

// Params fully describes what to render: which family, whether in Julia
// mode, where in the plane, and how hard to try before declaring a point
// inside the set.
//
// Scale is the width of the viewport in complex-plane units, so Params is
// independent of the pixel size of the window.
type Params struct {
	Type    string  `json:"type"`
	Julia   bool    `json:"julia,omitempty"`
	CenterX float64 `json:"center_x"`
	CenterY float64 `json:"center_y"`
	Scale   float64 `json:"scale"`
	MaxIter int     `json:"max_iter"`
	JuliaRe float64 `json:"julia_re"`
	JuliaIm float64 `json:"julia_im"`
}

// Default returns the classic opening view of the Mandelbrot set.
func Default() Params {
	return families[0].start()
}

// Unit returns the complex-plane size of one pixel for a viewport of width w.
func (p Params) Unit(w int) float64 {
	if w <= 0 {
		return p.Scale
	}
	return p.Scale / float64(w)
}

// PixelToPlane maps a pixel centre in a w x h viewport to a point in the plane.
// Pixel y grows downward, plane y grows upward.
func (p Params) PixelToPlane(px, py float64, w, h int) (x, y float64) {
	u := p.Unit(w)
	x = p.CenterX + (px-float64(w)/2+0.5)*u
	y = p.CenterY - (py-float64(h)/2+0.5)*u
	return x, y
}

// PlaneToPixel is the inverse of PixelToPlane.
func (p Params) PlaneToPixel(x, y float64, w, h int) (px, py float64) {
	u := p.Unit(w)
	px = (x-p.CenterX)/u + float64(w)/2 - 0.5
	py = (p.CenterY-y)/u + float64(h)/2 - 0.5
	return px, py
}

// Pan returns a copy shifted by a fraction of the viewport width and height.
func (p Params) Pan(fx, fy float64, w, h int) Params {
	u := p.Unit(w)
	p.CenterX += fx * float64(w) * u
	p.CenterY -= fy * float64(h) * u
	return p
}

// ZoomAt returns a copy zoomed by factor (<1 zooms in) keeping the plane
// point under pixel (px, py) fixed on screen.
func (p Params) ZoomAt(factor, px, py float64, w, h int) Params {
	x, y := p.PixelToPlane(px, py, w, h)
	p.Scale *= factor
	nx, ny := p.PixelToPlane(px, py, w, h)
	p.CenterX += x - nx
	p.CenterY += y - ny
	return p
}

// Zoom returns a copy zoomed by factor about the viewport centre.
func (p Params) Zoom(factor float64) Params {
	p.Scale *= factor
	return p
}

// Label is the human-readable name of the current fractal, for example
// "burningship julia".
func (p Params) Label() string {
	if p.Julia {
		return p.Type + " julia"
	}
	return p.Type
}

// Fractal computes the smooth escape value of a single point.
type Fractal interface {
	// Name is the family key, for example "mandelbrot".
	Name() string
	// Iterate returns a smooth iteration count >= 0 for points that escape
	// (or converge, for root finders) within maxIter iterations, or Inside
	// for points that do not.
	Iterate(x, y float64, maxIter int) float32
}

// Inside is the value Iterate returns for points that never escape.
const Inside float32 = -1
