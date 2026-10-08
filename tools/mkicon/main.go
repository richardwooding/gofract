// Command mkicon renders the application icon: a Julia set on a dark disc,
// written at several sizes for desktop integration. Run from the repo root:
//
//	go run ./tools/mkicon assets
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"

	"golang.org/x/image/draw"

	"github.com/richardwooding/gofract/internal/fractal"
	"github.com/richardwooding/gofract/internal/palette"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: mkicon OUTDIR")
	}
	dir := os.Args[1]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}
	const base = 1024
	src := render(base)
	for _, size := range []int{1024, 512, 256, 128, 64, 48, 32, 16} {
		dst := image.NewRGBA(image.Rect(0, 0, size, size))
		draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
		path := filepath.Join(dir, fmt.Sprintf("icon-%d.png", size))
		if err := write(path, dst); err != nil {
			log.Fatal(err)
		}
		fmt.Println(path)
	}
}

// render draws the Julia set for c = -0.8 + 0.156i inside a rounded square.
func render(n int) *image.RGBA {
	p := fractal.Default().ToJulia(-0.8, 0.156)
	p.Scale = 3.3
	p.MaxIter = 400
	f, err := fractal.New(p)
	if err != nil {
		log.Fatal(err)
	}
	pal := palette.Presets()[1] // fire
	img := image.NewRGBA(image.Rect(0, 0, n, n))
	bg := color.RGBA{14, 10, 28, 255}
	radius := float64(n) * 0.22
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			a := roundedAlpha(float64(x)+0.5, float64(y)+0.5, float64(n), radius)
			if a == 0 {
				continue
			}
			dx, dy := p.PixelOffset(float64(x), float64(y), n, n)
			v := f.Iterate(dx, dy, p.MaxIter)
			c := bg
			if v >= 0 {
				c = pal.Colors[1+int(v*2)%palette.Ring]
			}
			img.SetRGBA(x, y, color.RGBA{
				uint8(float64(c.R) * a), uint8(float64(c.G) * a), uint8(float64(c.B) * a), uint8(255 * a),
			})
		}
	}
	return img
}

// roundedAlpha returns coverage of a rounded square of side n with the
// given corner radius, with a one pixel antialiased edge.
func roundedAlpha(x, y, n, r float64) float64 {
	cx := math.Max(r, math.Min(n-r, x))
	cy := math.Max(r, math.Min(n-r, y))
	d := math.Hypot(x-cx, y-cy) - r
	switch {
	case d <= -0.5:
		return 1
	case d >= 0.5:
		return 0
	}
	return 0.5 - d
}

func write(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
