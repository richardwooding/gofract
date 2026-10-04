package render

import (
	"testing"

	"github.com/richardwooding/gofract/internal/fractal"
)

func renderOnce(t testing.TB, workers, w, h int) []float32 {
	t.Helper()
	r := New(workers)
	p := fractal.Default()
	r.Start(mandel(), p, w, h)
	r.Wait()
	if r.Busy() {
		t.Fatal("still busy after Wait")
	}
	if got := r.Pass(); got != Passes() {
		t.Fatalf("pass = %d, want %d", got, Passes())
	}
	buf, bw, bh := r.Snapshot(nil)
	if bw != w || bh != h || len(buf) != w*h {
		t.Fatalf("snapshot dims %dx%d len %d", bw, bh, len(buf))
	}
	return buf
}

func TestDeterministicAcrossWorkerCounts(t *testing.T) {
	a := renderOnce(t, 1, 203, 77) // odd sizes exercise edge blocks
	b := renderOnce(t, 8, 203, 77)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("pixel %d differs: %v vs %v", i, a[i], b[i])
		}
	}
}

func TestFinalPassMatchesDirectEvaluation(t *testing.T) {
	w, h := 160, 90
	buf := renderOnce(t, 4, w, h)
	p := fractal.Default()
	m := mandel()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx, dy := p.PixelOffset(float64(x), float64(y), w, h)
			want := m.Iterate(dx, dy, p.MaxIter)
			if got := buf[y*w+x]; got != want {
				t.Fatalf("(%d,%d) = %v, want %v", x, y, got, want)
			}
		}
	}
}

func TestRestartAbandonsOldRender(t *testing.T) {
	r := New(2)
	p := fractal.Default()
	p.MaxIter = 20000
	p.Scale = 1e-6 // deep enough to be slow
	r.Start(mandel(), p, 400, 400)
	q := fractal.Default()
	r.Start(mandel(), q, 64, 64)
	r.Wait()
	buf, w, h := r.Snapshot(nil)
	if w != 64 || h != 64 {
		t.Fatalf("dims %dx%d", w, h)
	}
	m := mandel()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx, dy := q.PixelOffset(float64(x), float64(y), w, h)
			if got, want := buf[y*w+x], m.Iterate(dx, dy, q.MaxIter); got != want {
				t.Fatalf("stale data at (%d,%d): %v want %v", x, y, got, want)
			}
		}
	}
}

func BenchmarkMandelbrot1280x720(b *testing.B) {
	r := New(0)
	p := fractal.Default()
	for i := 0; i < b.N; i++ {
		r.Start(mandel(), p, 1280, 720)
		r.Wait()
	}
}

func mandel() fractal.Fractal {
	f, err := fractal.New(fractal.Default())
	if err != nil {
		panic(err)
	}
	return f
}
