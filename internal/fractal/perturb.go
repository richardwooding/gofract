package fractal

import (
	"math/big"
	"sync"
)

// perturb renders the Mandelbrot set beyond float64 by perturbation: one
// reference orbit Z is computed at the view centre in arbitrary precision,
// and each pixel iterates only its small offset d from it in float64:
//
//	d' = 2 Z d + d^2 + dc
//
// When the full value z = Z + d becomes smaller than d the pixel has lost
// precision relative to the reference, so it rebases: d is replaced by z and
// the reference index restarts at zero (Z[0] = 0, so the new d is exact).
// The same rebase handles a reference that escapes before maxIter. This is
// Zhuoran's rebasing scheme; it needs no glitch detection passes.
type perturb struct {
	p    Params
	once sync.Once
	ref  []complex128
}

func newPerturb(p Params) *perturb { return &perturb{p: p} }

func (m *perturb) Name() string { return "mandelbrot" }

// prepare computes the reference orbit Z[0..n] at the centre, stopping when
// it escapes. Each Z is stored rounded to complex128; the rounding error is
// absorbed by the rebasing.
func (m *perturb) prepare() {
	prec := m.p.Prec()
	cx, cy := m.p.center()
	zx := new(big.Float).SetPrec(prec)
	zy := new(big.Float).SetPrec(prec)
	t1 := new(big.Float).SetPrec(prec)
	t2 := new(big.Float).SetPrec(prec)
	n := m.p.MaxIter
	if n < 1 {
		n = 1
	}
	ref := make([]complex128, 0, n+1)
	ref = append(ref, 0)
	for i := 0; i < n; i++ {
		// z = z^2 + c
		t1.Mul(zx, zx)
		t2.Mul(zy, zy)
		nx := new(big.Float).SetPrec(prec).Sub(t1, t2)
		nx.Add(nx, cx)
		t1.Mul(zx, zy)
		t1.Add(t1, t1)
		zy.Add(t1, cy)
		zx = nx
		fx, _ := zx.Float64()
		fy, _ := zy.Float64()
		ref = append(ref, complex(fx, fy))
		if fx*fx+fy*fy > bailout {
			break
		}
	}
	m.ref = ref
}

func (m *perturb) Iterate(dx, dy float64, maxIter int) float32 {
	m.once.Do(m.prepare)
	ref := m.ref
	last := len(ref) - 1
	dc := complex(dx, dy)
	var d complex128
	k := 0 // reference index
	for n := 0; n < maxIter; n++ {
		d = 2*ref[k]*d + d*d + dc
		k++
		z := ref[k] + d
		zz := real(z)*real(z) + imag(z)*imag(z)
		if zz > bailout {
			return smooth(n+1, zz, 2)
		}
		dd := real(d)*real(d) + imag(d)*imag(d)
		if zz < dd || k == last {
			d = z
			k = 0
		}
	}
	return Inside
}
