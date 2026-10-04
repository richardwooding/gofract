package fractal

import (
	"math"
	"testing"
)

// boundary is a point on the set boundary to float64 accuracy, plus a
// nearby point comfortably inside the set.
type boundary struct {
	x, y     float64
	inX, inY float64
}

// findBoundary scans rays out of (x0, y0) and bisects the first
// inside/outside transition on each, returning the first one accept
// approves of. Smooth arcs of a boundary (parabolic points) have escape
// times that grow like 1/sqrt(distance), so a window a few pixels wide
// around them looks flat at any sane iteration limit; accept lets a test
// skip those in favour of filament regions.
func findBoundary(t *testing.T, f Fractal, maxIter int, x0, y0 float64, accept func(boundary) bool) boundary {
	t.Helper()
	in := func(x, y float64) bool { return f.Iterate(x, y, maxIter) == Inside }
	for k := 0; k < 24; k++ {
		ang := 0.37 + float64(k)*math.Pi/12
		ux, uy := math.Cos(ang), math.Sin(ang)
		const steps, reach = 400, 2.5
		px, py := x0, y0
		pin := in(px, py)
		for i := 1; i <= steps; i++ {
			x := x0 + reach*ux*float64(i)/steps
			y := y0 + reach*uy*float64(i)/steps
			if in(x, y) == pin {
				px, py = x, y
				continue
			}
			ax, ay, bx, by := px, py, x, y // a is on the start side
			for j := 0; j < 60; j++ {
				mx, my := (ax+bx)/2, (ay+by)/2
				if in(mx, my) == pin {
					ax, ay = mx, my
				} else {
					bx, by = mx, my
				}
			}
			b := boundary{x: (ax + bx) / 2, y: (ay + by) / 2}
			// Step 0.01 toward the inside.
			dir := -0.01
			if !pin {
				dir = 0.01
			}
			b.inX, b.inY = b.x+dir*ux, b.y+dir*uy
			if !in(b.inX, b.inY) || !accept(b) {
				break // try the next ray
			}
			return b
		}
	}
	t.Fatalf("no boundary found around (%v,%v)", x0, y0)
	return boundary{}
}

// deepCases lists each perturbation family in Mandelbrot mode and in Julia
// mode with a constant chosen just inside the set, so the Julia set has an
// interior and the origin is inside it.
func deepCases(t *testing.T, maxIter int, accept func(Params) bool) []Params {
	var out []Params
	for name := range deepFamilies {
		p, err := Start(name)
		if err != nil {
			t.Fatal(err)
		}
		p.MaxIter = maxIter
		at := func(p Params, b boundary) Params {
			p.CenterX, p.CenterY = Dec(b.x), Dec(b.y)
			return p
		}
		cx, cy := p.CenterF()
		b := findBoundary(t, origin(t, p), maxIter, cx, cy, func(b boundary) bool { return accept(at(p, b)) })
		out = append(out, at(p, b))

		j := p.ToJulia(b.inX, b.inY)
		jb := findBoundary(t, origin(t, j), maxIter, 0, 0, func(b boundary) bool { return accept(at(j, b)) })
		out = append(out, at(j, jb))
	}
	return out
}

// structured reports whether a small perturbation render of p at the given
// scale shows both inside and outside points and a spread of escape counts.
func structured(p Params, scale float64) bool {
	p.Scale = scale
	f, err := New(p)
	if err != nil {
		return false
	}
	w, h := 16, 16
	vals := map[int]int{}
	inside := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx, dy := p.PixelOffset(float64(x), float64(y), w, h)
			v := f.Iterate(dx, dy, p.MaxIter)
			if v == Inside {
				inside++
			}
			vals[int(v)]++
		}
	}
	return inside > 0 && inside < w*h && len(vals) >= 5
}

func TestPerturbationFamiliesMatchFloat64(t *testing.T) {
	// At a zoom where float64 is still accurate, perturbation must agree
	// with the direct kernel as well as float64 agrees with itself: near a
	// chaotic boundary a one-ulp change of c flips inside/outside for many
	// pixels, so the budget for perturbation disagreement is set from that
	// measured conditioning rather than a fixed percentage.
	for _, p := range deepCases(t, 1500, func(Params) bool { return true }) {
		p.Scale = 1e-4
		direct, err := New(p)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := direct.(*perturb); ok {
			t.Fatalf("%s: expected direct kernel at scale %v", p.Label(), p.Scale)
		}
		nudged := p
		cx, _ := p.CenterF()
		nudged.CenterX = Dec(math.Nextafter(cx, 2))
		nd, err := New(nudged)
		if err != nil {
			t.Fatal(err)
		}
		pert := newPerturb(p, deepFamilies[p.Type])
		w, h := 64, 48
		mismatch, nudgeMismatch, inside := 0, 0, 0
		differs := func(a, b float32) bool {
			return (a == Inside) != (b == Inside) || math.Abs(float64(a-b)) > 0.01
		}
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dx, dy := p.PixelOffset(float64(x), float64(y), w, h)
				a := direct.Iterate(dx, dy, p.MaxIter)
				if a == Inside {
					inside++
				}
				if differs(a, pert.Iterate(dx, dy, p.MaxIter)) {
					mismatch++
				}
				if differs(a, nd.Iterate(dx, dy, p.MaxIter)) {
					nudgeMismatch++
				}
			}
		}
		if inside == 0 || inside == w*h {
			t.Errorf("%s: boundary region has %d inside of %d", p.Label(), inside, w*h)
		}
		budget := 2*nudgeMismatch + w*h/100
		if mismatch > budget {
			t.Errorf("%s: %d of %d pixels differ between perturbation and direct (one-ulp nudge changes %d)", p.Label(), mismatch, w*h, nudgeMismatch)
		}
	}
}

func TestPerturbationFamiliesDeep(t *testing.T) {
	// Regions are chosen where a 16x16 probe already shows structure; the
	// assertion below is on a larger window at the same depth.
	for _, p := range deepCases(t, 2000, func(p Params) bool { return structured(p, 1e-12) }) {
		p.Scale = 1e-12
		deep, err := New(p)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := deep.(*perturb); !ok {
			t.Fatalf("%s: expected perturbation at scale %v, got %T", p.Label(), p.Scale, deep)
		}
		w, h := 32, 32
		vals := map[int]int{}
		inside := 0
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dx, dy := p.PixelOffset(float64(x), float64(y), w, h)
				v := deep.Iterate(dx, dy, p.MaxIter)
				if v == Inside {
					inside++
				}
				vals[int(v)]++
			}
		}
		if inside == 0 || inside == w*h || len(vals) < 5 {
			t.Errorf("%s at %v: %d inside of %d, %d distinct values", p.Label(), p.Scale, inside, w*h, len(vals))
		}
	}
}

func TestDiffabs(t *testing.T) {
	for _, c := range []float64{-3, -1, -1e-9, 0, 1e-9, 1, 3} {
		for _, d := range []float64{-5, -1, -1e-12, 0, 1e-12, 1, 5} {
			want := math.Abs(c+d) - math.Abs(c)
			if got := diffabs(c, d); math.Abs(got-want) > 1e-15 {
				t.Errorf("diffabs(%v,%v) = %v, want %v", c, d, got, want)
			}
		}
	}
}
