package fractal

import "math"

// bailout is the squared magnitude beyond which an orbit is considered
// escaped. A large value makes the smooth colouring continuous.
const bailout = 1 << 16

// smooth converts the iteration index and squared magnitude at escape into
// a continuous iteration count: i + 1 - log2(log|z|), adjusted for the
// power of the map so the fractional part stays in [0, 1).
func smooth(i int, mag2, power float64) float32 {
	logMag := math.Log(mag2) / 2
	nu := math.Log(logMag/math.Ln2) / math.Log(power)
	return float32(float64(i) + 1 - nu)
}

// mandelbrot iterates z = z^2 + c.
func mandelbrot(zr, zi, cr, ci float64, maxIter int) float32 {
	for i := 0; i < maxIter; i++ {
		zr2 := zr * zr
		zi2 := zi * zi
		if zr2+zi2 > bailout {
			return smooth(i, zr2+zi2, 2)
		}
		zi = 2*zr*zi + ci
		zr = zr2 - zi2 + cr
	}
	return Inside
}

// inMainBulbs reports whether c lies in the main cardioid or the period-2
// bulb of the Mandelbrot set, the two largest regions that never escape.
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

// burningShip iterates z = (|Re z| + i|Im z|)^2 + c. The imaginary axis is
// negated so the ship appears the way it is usually pictured, hull down.
func burningShip(zr, zi, cr, ci float64, maxIter int) float32 {
	ci = -ci
	zi = -zi
	for i := 0; i < maxIter; i++ {
		zr2 := zr * zr
		zi2 := zi * zi
		if zr2+zi2 > bailout {
			return smooth(i, zr2+zi2, 2)
		}
		zi = math.Abs(2*zr*zi) + ci
		zr = zr2 - zi2 + cr
	}
	return Inside
}

// tricorn iterates z = conj(z)^2 + c, the Mandelbar set.
func tricorn(zr, zi, cr, ci float64, maxIter int) float32 {
	for i := 0; i < maxIter; i++ {
		zr2 := zr * zr
		zi2 := zi * zi
		if zr2+zi2 > bailout {
			return smooth(i, zr2+zi2, 2)
		}
		zi = -2*zr*zi + ci
		zr = zr2 - zi2 + cr
	}
	return Inside
}

// multibrot3 iterates z = z^3 + c.
func multibrot3(zr, zi, cr, ci float64, maxIter int) float32 {
	for i := 0; i < maxIter; i++ {
		zr2 := zr * zr
		zi2 := zi * zi
		if zr2+zi2 > bailout {
			return smooth(i, zr2+zi2, 3)
		}
		// (zr + i zi)^3 = zr^3 - 3 zr zi^2 + i (3 zr^2 zi - zi^3)
		nr := zr*zr2 - 3*zr*zi2 + cr
		ni := 3*zr2*zi - zi*zi2 + ci
		zr, zi = nr, ni
	}
	return Inside
}

// Newton's method on z^3 - 1, started from the pixel point. Points are coloured by the root they converge
// to (three bands a third of the palette apart) plus the iterations taken.
const (
	newtonEps    = 1e-12 // squared distance to a root that counts as converged
	newtonBand   = 85    // palette spacing between roots (255 / 3)
	newtonRootIm = 0.8660254037844386
)

func newton(_, _, zr, zi float64, maxIter int) float32 {
	for i := 0; i < maxIter; i++ {
		// Check convergence against the three cube roots of unity.
		if d := (zr-1)*(zr-1) + zi*zi; d < newtonEps {
			return newtonValue(i, d)
		}
		if d := (zr+0.5)*(zr+0.5) + (zi-newtonRootIm)*(zi-newtonRootIm); d < newtonEps {
			return newtonValue(i, d) + newtonBand
		}
		if d := (zr+0.5)*(zr+0.5) + (zi+newtonRootIm)*(zi+newtonRootIm); d < newtonEps {
			return newtonValue(i, d) + 2*newtonBand
		}
		// z - (z^3 - 1) / (3 z^2) = (2 z^3 + 1) / (3 z^2)
		zr2 := zr*zr - zi*zi
		zi2 := 2 * zr * zi
		den := 9 * (zr2*zr2 + zi2*zi2)
		if den == 0 {
			return Inside
		}
		// numerator 2 z^3 + 1 = 2 z * z^2 + 1
		nr := 2*(zr*zr2-zi*zi2) + 1
		ni := 2 * (zr*zi2 + zi*zr2)
		// divide by 3 z^2: (n * conj(3z^2)) / |3z^2|^2
		dr, di := 3*zr2, 3*zi2
		zr = (nr*dr + ni*di) / den
		zi = (ni*dr - nr*di) / den
	}
	return Inside
}

// newtonValue smooths the iteration count by how far inside the
// convergence radius the orbit landed; convergence is quadratic so the
// log-log ratio gives a fraction in [0, 1).
func newtonValue(i int, d float64) float32 {
	if d <= 0 {
		return float32(i)
	}
	f := math.Log(math.Log(d)/math.Log(newtonEps)) / math.Ln2
	if f < 0 {
		f = 0
	}
	if f > 1 {
		f = 1
	}
	return float32(float64(i) + 1 - f)
}
