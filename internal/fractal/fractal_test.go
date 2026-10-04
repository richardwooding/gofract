package fractal

import (
	"encoding/json"
	"math"
	"testing"
)

// origin builds a fractal with the view centred on 0 so Iterate takes
// absolute plane coordinates.
func origin(t *testing.T, p Params) Fractal {
	t.Helper()
	p.CenterX, p.CenterY = "0", "0"
	f, err := New(p)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestMandelbrotKnownPoints(t *testing.T) {
	m := origin(t, Default())
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
	m := origin(t, Default())
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
	j := origin(t, p)
	if got := j.Iterate(0.5, 0, 100); got != Inside {
		t.Errorf("|z|<1 with c=0 should be inside, got %v", got)
	}
	if got := j.Iterate(2, 0, 100); got < 0 {
		t.Errorf("|z|>1 with c=0 should escape, got %v", got)
	}
	back := p.ToMandelbrot()
	if back.Julia || back.CenterX.Float() != 0 || back.Scale != 3.5 {
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
		inside, outside := 0, 0
		for y := 0; y < 40; y++ {
			for x := 0; x < 40; x++ {
				dx, dy := p.PixelOffset(float64(x), float64(y), 40, 40)
				if f.Iterate(dx, dy, p.MaxIter) == Inside {
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
	f := origin(t, Params{Type: "newton", Scale: 4})
	v0 := f.Iterate(1, 0, 50)
	v1 := f.Iterate(-0.5, newtonRootIm, 50)
	v2 := f.Iterate(-0.5, -newtonRootIm, 50)
	if v0 < 0 || v0 >= 1 || int(v1) != newtonBand || int(v2) != 2*newtonBand {
		t.Errorf("root values %v %v %v", v0, v1, v2)
	}
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
	cx, cy := p.CenterF()
	if math.Abs(x-cx) > 1e-12 || math.Abs(y-cy) > 1e-12 {
		t.Errorf("centre pixel -> (%v,%v), want (%v,%v)", x, y, cx, cy)
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

func TestDeepZoomKeepsPrecision(t *testing.T) {
	// Zoom in 100 times by half at an off-centre pixel: the centre text must
	// grow to hold the extra digits, and recentring on the centre pixel must
	// be a no-op at full precision.
	p := Default()
	w, h := 800, 600
	for i := 0; i < 100; i++ {
		p = p.ZoomAt(0.5, 300, 200, w, h)
	}
	if p.Scale > 1e-29 || !p.Deep() {
		t.Fatalf("scale %v after 100 halvings", p.Scale)
	}
	if len(p.CenterX) < 30 {
		t.Errorf("centre lost digits: %q", p.CenterX)
	}
	q := p.Recenter(float64(w)/2-0.5, float64(h)/2-0.5, w, h)
	if q.CenterX != p.CenterX || q.CenterY != p.CenterY {
		t.Errorf("recentre on centre moved: %q -> %q", p.CenterX, q.CenterX)
	}
	// Panning right then left returns exactly.
	r := p.Pan(0.25, 0, w, h).Pan(-0.25, 0, w, h)
	if r.CenterX != p.CenterX {
		t.Errorf("pan round trip %q -> %q", p.CenterX, r.CenterX)
	}
}

func TestDecimalJSON(t *testing.T) {
	var p Params
	legacy := `{"type":"mandelbrot","center_x":-0.5,"center_y":0,"scale":3.5,"max_iter":256}`
	if err := json.Unmarshal([]byte(legacy), &p); err != nil {
		t.Fatal(err)
	}
	if p.CenterX != "-0.5" || p.CenterY != "0" {
		t.Errorf("legacy centre %q %q", p.CenterX, p.CenterY)
	}
	p.CenterX = "-0.74364388703715138230"
	out, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var q Params
	if err := json.Unmarshal(out, &q); err != nil {
		t.Fatal(err)
	}
	if q != p {
		t.Errorf("round trip %+v -> %+v", p, q)
	}
	if err := json.Unmarshal([]byte(`{"center_x":"abc"}`), &q); err == nil {
		t.Error("expected error for non-numeric centre")
	}
}

func TestPerturbationMatchesFloat64(t *testing.T) {
	// At zooms where float64 is still accurate, perturbation must agree
	// with the direct kernel pixel for pixel. The cardioid notch is almost
	// all inside; the seahorse valley mixes both and exercises rebasing.
	regions := []Params{
		{Type: "mandelbrot", CenterX: "0.25", CenterY: "0", Scale: 1e-4, MaxIter: 2000},
		{Type: "mandelbrot", CenterX: "-0.75", CenterY: "0.01", Scale: 1e-3, MaxIter: 2000},
		{Type: "mandelbrot", CenterX: "-0.7435", CenterY: "0.1314", Scale: 1e-5, MaxIter: 2000},
	}
	totalInside, total := 0, 0
	for ri, p := range regions {
		direct, err := New(p)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := direct.(*mandelSet); !ok {
			t.Fatalf("expected direct kernel at scale %v, got %T", p.Scale, direct)
		}
		pert := newPerturb(p)
		w, h := 64, 48
		mismatch, inside := 0, 0
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dx, dy := p.PixelOffset(float64(x), float64(y), w, h)
				a := direct.Iterate(dx, dy, p.MaxIter)
				b := pert.Iterate(dx, dy, p.MaxIter)
				if a == Inside {
					inside++
				}
				if (a == Inside) != (b == Inside) || math.Abs(float64(a-b)) > 0.01 {
					mismatch++
				}
			}
		}
		totalInside += inside
		total += w * h
		if mismatch > w*h/100 {
			t.Errorf("region %d: %d of %d pixels differ between perturbation and direct", ri, mismatch, w*h)
		}
	}
	if totalInside == 0 || totalInside == total {
		t.Fatalf("bad test regions: %d inside of %d", totalInside, total)
	}
}

func TestPerturbationDeep(t *testing.T) {
	// A known deep location: the image must have structure, not blocks.
	p := Params{Type: "mandelbrot", MaxIter: 20000, Scale: 1e-15,
		CenterX: "-0.743643887037158704752191506114774", CenterY: "0.131825904205311970493132056385139"}
	f, err := New(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := f.(*perturb); !ok {
		t.Fatalf("expected perturbation renderer, got %T", f)
	}
	w, h := 32, 32
	vals := map[int]int{}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx, dy := p.PixelOffset(float64(x), float64(y), w, h)
			vals[int(f.Iterate(dx, dy, p.MaxIter))]++
		}
	}
	if len(vals) < 20 {
		t.Errorf("only %d distinct values across %d pixels at scale %v", len(vals), w*h, p.Scale)
	}
}
