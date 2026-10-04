package fractal

// Mandelbrot is the set of c for which z = z^2 + c stays bounded from z = 0.
type Mandelbrot struct{}

// Name implements Fractal.
func (Mandelbrot) Name() string { return "mandelbrot" }

// Iterate implements Fractal.
func (Mandelbrot) Iterate(x, y float64, maxIter int) float32 {
	if inMainBulbs(x, y) {
		return Inside
	}
	return escape(0, 0, x, y, maxIter)
}

// inMainBulbs reports whether c lies in the main cardioid or the period-2
// bulb, the two largest regions that never escape. Skipping them avoids the
// most expensive full-maxIter orbits.
func inMainBulbs(x, y float64) bool {
	y2 := y * y
	xq := x - 0.25
	q := xq*xq + y2
	if q*(q+xq) <= 0.25*y2 {
		return true
	}
	xp := x + 1
	return xp*xp+y2 <= 0.0625
}
