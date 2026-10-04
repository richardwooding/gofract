package fractal

import (
	"math"
	"math/big"
	"sync"
)

// Perturbation rendering beyond float64.
//
// A reference orbit Z is computed in arbitrary precision and each pixel
// iterates only its small offset d from it in float64, using a recurrence
// specific to the family (for z^2 + c it is d' = 2 Z d + d^2 + dc). When the
// full value z = Z + d becomes smaller than d the pixel has lost precision
// relative to the reference, so it rebases: d is replaced by z and iteration
// continues against the critical orbit (the one starting at 0) from its
// first entry, which makes the new d exact. The same rebase handles a
// reference that escapes before maxIter. This is Zhuoran's rebasing scheme;
// it needs no glitch detection passes.
//
// In Mandelbrot mode the reference is the critical orbit of the centre and
// pixels differ by dc. In Julia mode every pixel shares c, so the first
// reference starts at the centre with d = pixel offset, and rebasing moves
// to the critical orbit of c.

// deepFamily is the perturbation form of one escape-time family.
type deepFamily struct {
	power float64
	// delta advances the pixel offset d given the reference value Z and
	// the parameter offset dc (zero in Julia mode).
	delta func(Z, d, dc complex128) complex128
	// step advances a high-precision reference orbit in place.
	step func(s *bigOrbit)
	// flipY means the family is computed in conjugated coordinates (the
	// Burning Ship is drawn hull down).
	flipY bool
}

var deepFamilies = map[string]*deepFamily{
	"mandelbrot":  {power: 2, delta: deltaMandelbrot, step: (*bigOrbit).mandelbrot},
	"burningship": {power: 2, delta: deltaBurningShip, step: (*bigOrbit).burningShip, flipY: true},
	"tricorn":     {power: 2, delta: deltaTricorn, step: (*bigOrbit).tricorn},
	"multibrot3":  {power: 3, delta: deltaMultibrot3, step: (*bigOrbit).multibrot3},
}

func deltaMandelbrot(Z, d, dc complex128) complex128 {
	return 2*Z*d + d*d + dc
}

func deltaTricorn(Z, d, dc complex128) complex128 {
	// conj(Z + d)^2 - conj(Z)^2
	cz, cd := complex(real(Z), -imag(Z)), complex(real(d), -imag(d))
	return 2*cz*cd + cd*cd + dc
}

func deltaMultibrot3(Z, d, dc complex128) complex128 {
	// (Z + d)^3 - Z^3
	return 3*Z*Z*d + 3*Z*d*d + d*d*d + dc
}

// diffabs returns |c + d| - |c| without cancellation.
func diffabs(c, d float64) float64 {
	if c >= 0 {
		if c+d >= 0 {
			return d
		}
		return -(2*c + d)
	}
	if c+d > 0 {
		return 2*c + d
	}
	return -d
}

func deltaBurningShip(Z, d, dc complex128) complex128 {
	// With |X+x| = A + a and |Y+y| = B + b where A = |X|, a = diffabs(X, x):
	// (A+a + i(B+b))^2 - (A + iB)^2
	X, Y := real(Z), imag(Z)
	a := diffabs(X, real(d))
	b := diffabs(Y, imag(d))
	A, B := math.Abs(X), math.Abs(Y)
	return complex(2*A*a+a*a-2*B*b-b*b+real(dc), 2*(A*b+B*a+a*b)+imag(dc))
}

// bigOrbit iterates one reference orbit in arbitrary precision.
type bigOrbit struct {
	zx, zy, cx, cy *big.Float
	t1, t2, t3     *big.Float
}

func newBigOrbit(prec uint, zx, zy, cx, cy *big.Float) *bigOrbit {
	n := func() *big.Float { return new(big.Float).SetPrec(prec) }
	return &bigOrbit{
		zx: n().Set(zx), zy: n().Set(zy), cx: n().Set(cx), cy: n().Set(cy),
		t1: n(), t2: n(), t3: n(),
	}
}

func (s *bigOrbit) float() complex128 {
	fx, _ := s.zx.Float64()
	fy, _ := s.zy.Float64()
	return complex(fx, fy)
}

// mandelbrot: z = z^2 + c
func (s *bigOrbit) mandelbrot() {
	s.t1.Mul(s.zx, s.zx)
	s.t2.Mul(s.zy, s.zy)
	s.t3.Mul(s.zx, s.zy)
	s.zx.Sub(s.t1, s.t2)
	s.zx.Add(s.zx, s.cx)
	s.zy.Add(s.t3, s.t3)
	s.zy.Add(s.zy, s.cy)
}

// tricorn: z = conj(z)^2 + c
func (s *bigOrbit) tricorn() {
	s.t1.Mul(s.zx, s.zx)
	s.t2.Mul(s.zy, s.zy)
	s.t3.Mul(s.zx, s.zy)
	s.zx.Sub(s.t1, s.t2)
	s.zx.Add(s.zx, s.cx)
	s.zy.Add(s.t3, s.t3)
	s.zy.Neg(s.zy)
	s.zy.Add(s.zy, s.cy)
}

// burningShip: z = (|x| + i|y|)^2 + c
func (s *bigOrbit) burningShip() {
	s.zx.Abs(s.zx)
	s.zy.Abs(s.zy)
	s.mandelbrot()
}

// multibrot3: z = z^3 + c
func (s *bigOrbit) multibrot3() {
	s.t1.Mul(s.zx, s.zx) // x^2
	s.t2.Mul(s.zy, s.zy) // y^2
	// nx = x^3 - 3 x y^2 + cx
	nx := new(big.Float).SetPrec(s.zx.Prec()).Mul(s.t1, s.zx)
	s.t3.Mul(s.zx, s.t2)
	s.t3.Add(s.t3, s.t3)
	s.t3.Add(s.t3, new(big.Float).SetPrec(s.zx.Prec()).Mul(s.zx, s.t2))
	nx.Sub(nx, s.t3)
	nx.Add(nx, s.cx)
	// ny = 3 x^2 y - y^3 + cy
	ny := new(big.Float).SetPrec(s.zx.Prec()).Mul(s.t1, s.zy)
	s.t3.Add(ny, ny)
	ny.Add(ny, s.t3)
	s.t3.Mul(s.t2, s.zy)
	ny.Sub(ny, s.t3)
	ny.Add(ny, s.cy)
	s.zx, s.zy = nx, ny
}

// run iterates up to n steps, returning Z[0..k] where Z[k] is the first
// escaped value or Z[n].
func (s *bigOrbit) run(step func(*bigOrbit), n int) []complex128 {
	ref := make([]complex128, 0, n+1)
	ref = append(ref, s.float())
	for i := 0; i < n; i++ {
		step(s)
		z := s.float()
		ref = append(ref, z)
		if real(z)*real(z)+imag(z)*imag(z) > bailout {
			break
		}
	}
	return ref
}

type perturb struct {
	p    Params
	fam  *deepFamily
	once sync.Once
	main []complex128 // reference the pixel starts against
	crit []complex128 // critical orbit used after a rebase
}

func newPerturb(p Params, fam *deepFamily) *perturb { return &perturb{p: p, fam: fam} }

func (m *perturb) Name() string { return m.p.Type }

func (m *perturb) prepare() {
	prec := m.p.Prec()
	n := m.p.MaxIter
	if n < 1 {
		n = 1
	}
	cx, cy := m.p.center()
	zero := new(big.Float).SetPrec(prec)
	if m.p.Julia {
		jx := new(big.Float).SetPrec(prec).SetFloat64(m.p.JuliaRe)
		jy := new(big.Float).SetPrec(prec).SetFloat64(m.p.JuliaIm)
		if m.fam.flipY {
			cy.Neg(cy)
			jy.Neg(jy)
		}
		m.main = newBigOrbit(prec, cx, cy, jx, jy).run(m.fam.step, n)
		m.crit = newBigOrbit(prec, zero, zero, jx, jy).run(m.fam.step, n)
		return
	}
	if m.fam.flipY {
		cy.Neg(cy)
	}
	m.main = newBigOrbit(prec, zero, zero, cx, cy).run(m.fam.step, n)
	m.crit = m.main
}

func (m *perturb) Iterate(dx, dy float64, maxIter int) float32 {
	m.once.Do(m.prepare)
	if m.fam.flipY {
		dy = -dy
	}
	var d, dc complex128
	if m.p.Julia {
		d = complex(dx, dy)
	} else {
		dc = complex(dx, dy)
	}
	delta := m.fam.delta
	ref := m.main
	last := len(ref) - 1
	k := 0
	for n := 0; n < maxIter; n++ {
		d = delta(ref[k], d, dc)
		k++
		z := ref[k] + d
		zz := real(z)*real(z) + imag(z)*imag(z)
		if zz > bailout {
			return smooth(n+1, zz, m.fam.power)
		}
		dd := real(d)*real(d) + imag(d)*imag(d)
		if zz < dd || k == last {
			d = z
			ref = m.crit
			last = len(ref) - 1
			k = 0
		}
	}
	return Inside
}
