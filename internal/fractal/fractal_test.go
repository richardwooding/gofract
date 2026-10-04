package fractal

import (
	"math"
	"testing"
)

func mandel(t *testing.T) Fractal {
	t.Helper()
	f, err := New(Default())
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestMandelbrotKnownPoints(t *testing.T) {
	m := mandel(t)
	if got := m.Iterate(0, 0, 100); got != Inside {
		t.Errorf("origin should be inside, got %v", got)
	}
	if got := m.Iterate(-1, 0, 100); got != Inside {
		t.Errorf("-1 should be inside (period-2 bulb), got %v", got)
	}
	if got := m.Iterate(2, 2, 100); got < 0 || got > 3 {
		t.Errorf("2+2i should escape almost immediately, got %v", got)
	}
	if got := m.Iterate(0.26, 0, 1000); got < 10 {
		t.Errorf("0.26 should escape after many iterations, got %v", got)
	}
}

func TestBulbCheckAgreesWithIteration(t *testing.T) {
	for x := -1.6; x <= 0.6; x += 0.01 {
		for y := -1.2; y <= 1.2; y += 0.01 {
			if inMainBulbs(x, y) && mandelbrot(0, 0, x, y, 2000) != Inside {
				t.Fatalf("bulb check wrongly marked (%v,%v) inside", x, y)
			}
		}
	}
}

func TestSmoothValueIsMonotonicNearBoundary(t *testing.T) {
	m := mandel(t)
	prev := float32(math.Inf(1))
	for x := 0.26; x < 2; x += 0.05 {
		v := m.Iterate(x, 0, 500)
		if v < 0 || v > prev {
			t.Fatalf("expected decreasing escape counts, at x=%v got %v after %v", x, v, prev)
		}
		prev = v
	}
}

func TestJuliaMode(t *testing.T) {
	p := Default().ToJulia(0, 0)
	if !p.Julia || p.Type != "mandelbrot" {
		t.Fatalf("ToJulia gave %+v", p)
	}
	j, err := New(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := j.Iterate(0.5, 0, 100); got != Inside {
		t.Errorf("|z|<1 with c=0 should be inside, got %v", got)
	}
	if got := j.Iterate(2, 0, 100); got < 0 {
		t.Errorf("|z|>1 with c=0 should escape, got %v", got)
	}
	back := p.ToMandelbrot()
	if back.Julia || back.CenterX != 0 || back.Scale != 3.5 {
		t.Errorf("ToMandelbrot gave %+v", back)
	}
}

func TestLegacyJuliaType(t *testing.T) {
	f, err := New(Params{Type: "julia", JuliaRe: -0.8, JuliaIm: 0.156})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := f.(*juliaSet); !ok {
		t.Errorf("legacy julia type built %T", f)
	}
}

func TestEveryFamilyRendersSomething(t *testing.T) {
	for _, name := range Names() {
		p, err := Start(name)
		if err != nil {
			t.Fatal(err)
		}
		f, err := New(p)
		if err != nil {
			t.Fatalf("New(%q): %v", name, err)
		}
		if f.Name() != name {
			t.Errorf("New(%q).Name() = %q", name, f.Name())
		}
		// Sample a grid over the default view; expect a mix of inside and
		// outside points so the picture is not blank.
		inside, outside := 0, 0
		for y := 0; y < 40; y++ {
			for x := 0; x < 40; x++ {
				px, py := p.PixelToPlane(float64(x), float64(y), 40, 40)
				if f.Iterate(px, py, p.MaxIter) == Inside {
					inside++
				} else {
					outside++
				}
			}
		}
		if outside == 0 {
			t.Errorf("%s: every sample inside", name)
		}
		if inside == 0 && name != "newton" {
			t.Errorf("%s: every sample outside", name)
		}
		if HasJulia(name) {
			if _, err := New(p.ToJulia(p.JuliaRe, p.JuliaIm)); err != nil {
				t.Errorf("%s julia: %v", name, err)
			}
		} else if q := p.ToJulia(0, 0); q.Julia {
			t.Errorf("%s: ToJulia should be a no-op", name)
		}
	}
	if _, err := New(Params{Type: "nope"}); err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestNewtonRoots(t *testing.T) {
	f, _ := New(Params{Type: "newton"})
	// Starting on a root converges immediately into that root's band.
	v0 := f.Iterate(1, 0, 50)
	v1 := f.Iterate(-0.5, newtonRootIm, 50)
	v2 := f.Iterate(-0.5, -newtonRootIm, 50)
	if v0 < 0 || v0 >= 1 || int(v1) != newtonBand || int(v2) != 2*newtonBand {
		t.Errorf("root values %v %v %v", v0, v1, v2)
	}
	// The real axis right of the origin converges to root 1.
	if v := f.Iterate(2, 0, 50); v < 1 || v >= newtonBand {
		t.Errorf("2+0i -> %v, want root 1 band", v)
	}
	if v := f.Iterate(0, 0, 50); v != Inside {
		t.Errorf("origin has zero derivative, got %v", v)
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
