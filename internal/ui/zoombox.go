package ui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/richardwooding/gofract/internal/fractal"
)

// minBox is the smallest drag, in pixels, that counts as a zoom rather than
// a click.
const minBox = 6

// zoomBox is a rubber-band rectangle anchored where the drag began and
// locked to the viewport's aspect ratio.
type zoomBox struct {
	active bool
	ax, ay int
	cx, cy int
}

func (z *zoomBox) begin(x, y int) {
	z.active = true
	z.ax, z.ay = x, y
	z.cx, z.cy = x, y
}

func (z *zoomBox) move(x, y int) { z.cx, z.cy = x, y }

// rect returns the aspect-locked box as top-left corner and size.
func (z *zoomBox) rect(w, h int) (x, y, bw, bh float64) {
	dx := float64(z.cx - z.ax)
	dy := float64(z.cy - z.ay)
	aspect := float64(w) / float64(h)
	bw = math.Max(math.Abs(dx), math.Abs(dy)*aspect)
	bh = bw / aspect
	x = float64(z.ax)
	y = float64(z.ay)
	if dx < 0 {
		x -= bw
	}
	if dy < 0 {
		y -= bh
	}
	return x, y, bw, bh
}

// commit turns the box into new view parameters. ok is false for drags too
// small to be deliberate.
func (z *zoomBox) commit(p fractal.Params, w, h int) (fractal.Params, bool) {
	x, y, bw, bh := z.rect(w, h)
	if bw < minBox || bh < minBox {
		return p, false
	}
	p = p.Recenter(x+bw/2, y+bh/2, w, h)
	p.Scale *= bw / float64(w)
	return p, true
}

func (z *zoomBox) draw(dst *ebiten.Image, w, h int) {
	if !z.active {
		return
	}
	x, y, bw, bh := z.rect(w, h)
	if bw < 1 || bh < 1 {
		return
	}
	vector.StrokeRect(dst, float32(x)-1, float32(y)-1, float32(bw)+2, float32(bh)+2, 1, color.RGBA{0, 0, 0, 200}, false)
	vector.StrokeRect(dst, float32(x), float32(y), float32(bw), float32(bh), 1, color.RGBA{255, 255, 255, 255}, false)
}
