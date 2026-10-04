package fractal

import (
	"math"
	"testing"
)

func TestMandelbrotKnownPoints(t *testing.T) {
	m := Mandelbrot{}
	if got := m.Iterate(0, 0, 100); got != Inside {
		t.Errorf("origin should be inside, got %v", got)
	}
	if got := m.Iterate(-1, 0, 100); got != Inside {
		t.Errorf("-1 should be inside (period-2 bulb), got %v", got)
	}
	if got := m.Iterate(2, 2, 100); got < 0 || got > 3 {
		t.Errorf("2+2i should escape almost immediately, got %v", got)
	}
	// A point just outside the cardioid cusp escapes late but does escape.
	if got := m.Iterate(0.26, 0, 1000); got < 10 {
		t.Errorf("0.26 should escape after many iterations, got %v", got)
	}
}

func TestBulbCheckAgreesWithIteration(t *testing.T) {
	// Every point the shortcut claims is inside must also fail to escape.
	for x := -1.6; x <= 0.6; x += 0.01 {
		for y := -1.2; y <= 1.2; y += 0.01 {
			if inMainBulbs(x, y) && escape(0, 0, x, y, 2000) != Inside {
				t.Fatalf("bulb check wrongly marked (%v,%v) inside", x, y)
			}
		}
	}
}

func TestSmoothValueIsMonotonicNearBoundary(t *testing.T) {
	// Moving away from the set along the real axis, the escape count falls.
	m := Mandelbrot{}
	prev := float32(math.Inf(1))
	for x := 0.26; x < 2; x += 0.05 {
		v := m.Iterate(x, 0, 500)
		if v < 0 || v > prev {
			t.Fatalf("expected decreasing escape counts, at x=%v got %v after %v", x, v, prev)
		}
		prev = v
	}
}

func TestJulia(t *testing.T) {
	j := Julia{C: 0}
	if got := j.Iterate(0.5, 0, 100); got != Inside {
		t.Errorf("|z|<1 with c=0 should be inside, got %v", got)
	}
	if got := j.Iterate(2, 0, 100); got < 0 {
		t.Errorf("|z|>1 with c=0 should escape, got %v", got)
	}
}

func TestRegistry(t *testing.T) {
	for _, name := range Names() {
		f, err := New(Params{Type: name})
		if err != nil {
			t.Fatalf("New(%q): %v", name, err)
		}
		if f.Name() != name {
			t.Errorf("New(%q).Name() = %q", name, f.Name())
		}
	}
	if _, err := New(Params{Type: "nope"}); err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestPixelPlaneRoundTrip(t *testing.T) {
	p := Default()
	w, h := 1280, 720
	for _, px := range []float64{0, 100.5, 639.5, 1279} {
		for _, py := range []float64{0, 359.5, 719} {
			x, y := p.PixelToPlane(px, py, w, h)
			bx, by := p.PlaneToPixel(x, y, w, h)
			if math.Abs(bx-px) > 1e-9 || math.Abs(by-py) > 1e-9 {
				t.Errorf("round trip (%v,%v) -> (%v,%v)", px, py, bx, by)
			}
		}
	}
	// Centre pixel maps to the centre of the view.
	x, y := p.PixelToPlane(float64(w)/2-0.5, float64(h)/2-0.5, w, h)
	if math.Abs(x-p.CenterX) > 1e-12 || math.Abs(y-p.CenterY) > 1e-12 {
		t.Errorf("centre pixel -> (%v,%v), want (%v,%v)", x, y, p.CenterX, p.CenterY)
	}
}

func TestZoomAtKeepsPointFixed(t *testing.T) {
	p := Default()
	w, h := 800, 600
	px, py := 123.0, 456.0
	x0, y0 := p.PixelToPlane(px, py, w, h)
	z := p.ZoomAt(0.5, px, py, w, h)
	x1, y1 := z.PixelToPlane(px, py, w, h)
	if math.Abs(x0-x1) > 1e-12 || math.Abs(y0-y1) > 1e-12 {
		t.Errorf("point under cursor moved from (%v,%v) to (%v,%v)", x0, y0, x1, y1)
	}
	if z.Scale != p.Scale/2 {
		t.Errorf("scale = %v, want %v", z.Scale, p.Scale/2)
	}
}
