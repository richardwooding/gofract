package fractal

// Julia is the set of z0 for which z = z^2 + C stays bounded for a fixed C.
type Julia struct {
	C complex128
}

// Name implements Fractal.
func (Julia) Name() string { return "julia" }

// Iterate implements Fractal.
func (j Julia) Iterate(x, y float64, maxIter int) float32 {
	return escape(x, y, real(j.C), imag(j.C), maxIter)
}
