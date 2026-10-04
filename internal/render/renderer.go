// Package render computes fractal iteration buffers on a pool of goroutines,
// filling the image progressively from coarse to fine the way Fractint's
// solid-guessing pass did.
package render

import (
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/richardwooding/gofract/internal/fractal"
)

// strides are the pixel steps of the progressive passes, coarsest first.
var strides = []int{8, 4, 2, 1}

// bandRows is the height of one work unit. It is a multiple of the coarsest
// stride so every band starts on a sample point.
const bandRows = 32

// Renderer owns a float32 iteration buffer and the goroutines filling it.
// Only one render is live at a time; starting a new one abandons the old.
type Renderer struct {
	workers int

	mu     sync.Mutex
	gen    uint64
	width  int
	height int
	iters  []float32

	// Per-render bookkeeping for the live generation.
	cancel *atomic.Bool
	done   chan struct{}
	pass   atomic.Int32
	busy   atomic.Bool
}

// New returns a renderer using the given number of worker goroutines, or
// runtime.NumCPU() when workers <= 0.
func New(workers int) *Renderer {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	done := make(chan struct{})
	close(done)
	return &Renderer{workers: workers, cancel: new(atomic.Bool), done: done}
}

// Passes returns the number of progressive passes a render goes through.
func Passes() int { return len(strides) }

// Start abandons any in-flight render and begins computing f at p into a
// w x h buffer. If the size is unchanged the previous image stays visible
// until the first coarse pass overwrites it.
func (r *Renderer) Start(f fractal.Fractal, p fractal.Params, w, h int) {
	r.mu.Lock()
	r.cancel.Store(true)
	r.gen++
	gen := r.gen
	if w != r.width || h != r.height {
		r.width, r.height = w, h
		r.iters = make([]float32, w*h)
		for i := range r.iters {
			r.iters[i] = fractal.Inside
		}
	}
	cancel := new(atomic.Bool)
	done := make(chan struct{})
	r.cancel = cancel
	r.done = done
	r.pass.Store(0)
	r.busy.Store(true)
	r.mu.Unlock()

	if w <= 0 || h <= 0 {
		r.busy.Store(false)
		close(done)
		return
	}
	go r.run(gen, cancel, done, f, p, w, h)
}

// Stop abandons the in-flight render, if any.
func (r *Renderer) Stop() {
	r.mu.Lock()
	r.cancel.Store(true)
	r.mu.Unlock()
}

// Wait blocks until the most recently started render finishes or is
// abandoned.
func (r *Renderer) Wait() {
	r.mu.Lock()
	done := r.done
	r.mu.Unlock()
	<-done
}

// Busy reports whether a render is in progress.
func (r *Renderer) Busy() bool { return r.busy.Load() }

// Pass returns the number of completed progressive passes of the live
// render, from 0 to Passes().
func (r *Renderer) Pass() int { return int(r.pass.Load()) }

// Snapshot copies the current buffer into dst, growing it if needed, and
// returns the slice together with its dimensions.
func (r *Renderer) Snapshot(dst []float32) ([]float32, int, int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := r.width * r.height
	if cap(dst) < n {
		dst = make([]float32, n)
	}
	dst = dst[:n]
	copy(dst, r.iters)
	return dst, r.width, r.height
}

// band is one unit of work: rows [y0, y1) at a given stride.
type band struct {
	y0, y1 int
	stride int
	first  bool // first pass: no earlier samples to skip
}

func (r *Renderer) run(gen uint64, cancel *atomic.Bool, done chan struct{}, f fractal.Fractal, p fractal.Params, w, h int) {
	defer close(done)
	defer r.busy.Store(false)

	for pi, s := range strides {
		if cancel.Load() {
			return
		}
		jobs := make(chan band)
		var wg sync.WaitGroup
		for i := 0; i < r.workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				local := make([]float32, 0, bandRows*w)
				for b := range jobs {
					if cancel.Load() {
						continue
					}
					local = r.renderBand(local[:0], cancel, f, p, w, h, b)
					if local != nil {
						r.commit(gen, b, w, local)
					}
				}
			}()
		}
		for y := 0; y < h; y += bandRows {
			y1 := y + bandRows
			if y1 > h {
				y1 = h
			}
			jobs <- band{y0: y, y1: y1, stride: s, first: pi == 0}
		}
		close(jobs)
		wg.Wait()
		if cancel.Load() {
			return
		}
		r.pass.Store(int32(pi + 1))
	}
}

// renderBand computes every stride-aligned sample in the band and expands it
// into a stride x stride block. The result is a dense (y1-y0) x w slice, or
// nil if the render was cancelled part way.
func (r *Renderer) renderBand(buf []float32, cancel *atomic.Bool, f fractal.Fractal, p fractal.Params, w, h int, b band) []float32 {
	rows := b.y1 - b.y0
	buf = buf[:rows*w]
	s := b.stride
	unit := p.Unit(w)
	x0 := p.CenterX + (-float64(w)/2+0.5)*unit
	y0 := p.CenterY + (float64(h)/2-0.5)*unit
	maxIter := p.MaxIter
	if maxIter < 1 {
		maxIter = 1
	}
	// Skip samples already computed by the previous, coarser pass: they sit
	// on the 2s grid. When skipping we leave the block value as the
	// previously computed one, so fill the block from the existing buffer.
	skip := !b.first
	count := 0
	for y := b.y0; y < b.y1; y += s {
		py := y0 - float64(y)*unit
		for x := 0; x < w; x += s {
			count++
			if count&31 == 0 && cancel.Load() {
				return nil
			}
			var v float32
			if skip && x%(2*s) == 0 && y%(2*s) == 0 {
				v = r.peek(x, y)
			} else {
				v = f.Iterate(x0+float64(x)*unit, py, maxIter)
			}
			r.fill(buf, w, x, y-b.y0, s, rows, v)
		}
	}
	return buf
}

// fill writes v into the s x s block at (x, y) of a rows x w buffer.
func (r *Renderer) fill(buf []float32, w, x, y, s, rows int, v float32) {
	for dy := 0; dy < s && y+dy < rows; dy++ {
		row := buf[(y+dy)*w:]
		for dx := 0; dx < s && x+dx < w; dx++ {
			row[x+dx] = v
		}
	}
}

// peek reads one sample from the shared buffer. The value was committed by
// the previous pass of the same generation.
func (r *Renderer) peek(x, y int) float32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if y*r.width+x >= len(r.iters) {
		return fractal.Inside
	}
	return r.iters[y*r.width+x]
}

// commit copies a finished band into the shared buffer unless a newer render
// has taken over.
func (r *Renderer) commit(gen uint64, b band, w int, data []float32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if gen != r.gen {
		return
	}
	copy(r.iters[b.y0*w:b.y1*w], data)
}
