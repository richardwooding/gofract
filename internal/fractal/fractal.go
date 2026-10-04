// Package fractal defines the fractal families gofract can render and the
// view parameters that select a region of the complex plane.
package fractal

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// Decimal is an arbitrary-precision real stored as its decimal text, so a
// deep-zoom centre survives JSON and value comparison. It also accepts a
// plain JSON number for sidecars written before deep zoom existed.
type Decimal string

// Dec converts a float64 to a Decimal.
func Dec(f float64) Decimal { return Decimal(strconv.FormatFloat(f, 'g', -1, 64)) }

func decBig(f *big.Float) Decimal { return Decimal(f.Text('g', -1)) }

// Float returns the nearest float64.
func (d Decimal) Float() float64 {
	if d == "" {
		return 0
	}
	f, _, err := big.ParseFloat(string(d), 10, 64, big.ToNearestEven)
	if err != nil {
		return 0
	}
	v, _ := f.Float64()
	return v
}

// Big parses the value at the given precision in bits.
func (d Decimal) Big(prec uint) *big.Float {
	if d == "" {
		return new(big.Float).SetPrec(prec)
	}
	f, _, err := big.ParseFloat(string(d), 10, prec, big.ToNearestEven)
	if err != nil {
		return new(big.Float).SetPrec(prec)
	}
	return f
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Decimal) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "null" {
		*d = ""
		return nil
	}
	if strings.HasPrefix(s, `"`) {
		if err := json.Unmarshal(b, (*string)(d)); err != nil {
			return err
		}
		s = string(*d)
	}
	if _, _, err := big.ParseFloat(s, 10, 64, big.ToNearestEven); err != nil {
		return fmt.Errorf("decimal %q: %w", s, err)
	}
	*d = Decimal(s)
	return nil
}

// Params fully describes what to render: which family, whether in Julia
// mode, where in the plane, and how hard to try before declaring a point
// inside the set.
//
// Scale is the width of the viewport in complex-plane units, so Params is
// independent of the pixel size of the window. The centre is kept as text
// at whatever precision the zoom needs.
type Params struct {
	Type    string  `json:"type"`
	Julia   bool    `json:"julia,omitempty"`
	CenterX Decimal `json:"center_x"`
	CenterY Decimal `json:"center_y"`
	Scale   float64 `json:"scale"`
	MaxIter int     `json:"max_iter"`
	JuliaRe float64 `json:"julia_re"`
	JuliaIm float64 `json:"julia_im"`
}

// Default returns the classic opening view of the Mandelbrot set.
func Default() Params {
	return families[0].start()
}

// deepScale is the view width below which float64 pixel coordinates stop
// being distinct and the perturbation renderer takes over.
const deepScale = 1e-8

// Deep reports whether the view needs arbitrary precision.
func (p Params) Deep() bool { return p.Scale > 0 && p.Scale < deepScale }

// Prec is the number of mantissa bits needed for centre arithmetic at this
// zoom: enough to place a pixel with 64 bits to spare.
func (p Params) Prec() uint {
	bits := 64.0
	if p.Scale > 0 {
		bits -= math.Log2(p.Scale)
	}
	if bits < 64 {
		bits = 64
	}
	return uint(bits) + 64
}

// Unit returns the complex-plane size of one pixel for a viewport of width w.
func (p Params) Unit(w int) float64 {
	if w <= 0 {
		return p.Scale
	}
	return p.Scale / float64(w)
}

// CenterF returns the centre rounded to float64.
func (p Params) CenterF() (x, y float64) { return p.CenterX.Float(), p.CenterY.Float() }

func (p Params) center() (x, y *big.Float) {
	prec := p.Prec()
	return p.CenterX.Big(prec), p.CenterY.Big(prec)
}

func (p Params) withCenter(x, y *big.Float) Params {
	p.CenterX, p.CenterY = decBig(x), decBig(y)
	return p
}

// CenterText formats the centre with the given number of decimals.
func (p Params) CenterText(decimals int) (x, y string) {
	cx, cy := p.center()
	return cx.Text('f', decimals), cy.Text('f', decimals)
}

// PixelOffset returns the plane offset of a pixel centre from the view
// centre. This is exact in float64 at any zoom because it never involves the
// centre itself. Pixel y grows downward, plane y grows upward.
func (p Params) PixelOffset(px, py float64, w, h int) (dx, dy float64) {
	u := p.Unit(w)
	return (px - float64(w)/2 + 0.5) * u, -(py - float64(h)/2 + 0.5) * u
}

// PixelToPlane maps a pixel to an approximate float64 plane point. Use it
// for display and for picking Julia constants; use Recenter to navigate.
func (p Params) PixelToPlane(px, py float64, w, h int) (x, y float64) {
	dx, dy := p.PixelOffset(px, py, w, h)
	cx, cy := p.CenterF()
	return cx + dx, cy + dy
}

// PlaneToPixel is the float64 inverse of PixelToPlane.
func (p Params) PlaneToPixel(x, y float64, w, h int) (px, py float64) {
	u := p.Unit(w)
	cx, cy := p.CenterF()
	px = (x-cx)/u + float64(w)/2 - 0.5
	py = (cy-y)/u + float64(h)/2 - 0.5
	return px, py
}

func addF(x *big.Float, d float64) *big.Float {
	t := new(big.Float).SetPrec(x.Prec()).SetFloat64(d)
	return new(big.Float).SetPrec(x.Prec()).Add(x, t)
}

// Recenter moves the view centre to the plane point under pixel (px, py).
func (p Params) Recenter(px, py float64, w, h int) Params {
	dx, dy := p.PixelOffset(px, py, w, h)
	cx, cy := p.center()
	return p.withCenter(addF(cx, dx), addF(cy, dy))
}

// Pan returns a copy shifted by a fraction of the viewport width and height.
func (p Params) Pan(fx, fy float64, w, h int) Params {
	u := p.Unit(w)
	cx, cy := p.center()
	return p.withCenter(addF(cx, fx*float64(w)*u), addF(cy, -fy*float64(h)*u))
}

// ZoomAt returns a copy zoomed by factor (<1 zooms in) keeping the plane
// point under pixel (px, py) fixed on screen.
func (p Params) ZoomAt(factor, px, py float64, w, h int) Params {
	dx, dy := p.PixelOffset(px, py, w, h)
	cx, cy := p.center()
	x, y := addF(cx, dx), addF(cy, dy) // fixed point
	p.Scale *= factor
	ndx, ndy := p.PixelOffset(px, py, w, h)
	q := p.withCenter(addF(x, -ndx), addF(y, -ndy))
	return q
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

// Fractal computes the smooth escape value of a single pixel, given its
// plane offset from the view centre. Binding the centre at construction
// lets deep-zoom implementations keep it at arbitrary precision.
type Fractal interface {
	// Name is the family key, for example "mandelbrot".
	Name() string
	// Iterate returns a smooth iteration count >= 0 for points that escape
	// (or converge, for root finders) within maxIter iterations, or Inside
	// for points that do not.
	Iterate(dx, dy float64, maxIter int) float32
}

// Inside is the value Iterate returns for points that never escape.
const Inside float32 = -1
