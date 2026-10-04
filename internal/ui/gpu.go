package ui

import (
	_ "embed"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/richardwooding/gofract/internal/fractal"
	"github.com/richardwooding/gofract/internal/palette"
)

//go:embed fractal.kage
var kageSrc []byte

// gpuMinUnit is the smallest plane-units-per-pixel the float32 shader can
// resolve cleanly. Below it the CPU renderer takes over automatically.
const gpuMinUnit = 2e-6

// gpuMaxIter caps the shader loop; the Kage source unrolls to this bound.
const gpuMaxIter = 4096

type gpuMode int

const (
	gpuAuto gpuMode = iota
	gpuOff
	gpuOn
)

func (m gpuMode) String() string {
	switch m {
	case gpuOff:
		return "off"
	case gpuOn:
		return "on"
	}
	return "auto"
}

// gpu draws fractals with a Kage shader, colouring through a 256x1 palette
// texture so cycling costs nothing.
type gpu struct {
	shader   *ebiten.Shader
	pal      *ebiten.Image
	palSrc   *palette.Palette
	vertices []ebiten.Vertex
	indices  []uint16
}

func newGPU() (*gpu, error) {
	s, err := ebiten.NewShader(kageSrc)
	if err != nil {
		return nil, fmt.Errorf("compile fractal shader: %w", err)
	}
	g := &gpu{
		shader:   s,
		pal:      ebiten.NewImage(palette.Size, 1),
		vertices: make([]ebiten.Vertex, 4),
		indices:  []uint16{0, 1, 2, 1, 3, 2},
	}
	for i := range g.vertices {
		g.vertices[i].ColorR, g.vertices[i].ColorG, g.vertices[i].ColorB, g.vertices[i].ColorA = 1, 1, 1, 1
	}
	return g, nil
}

// canRender reports whether the shader can draw p faithfully.
func (g *gpu) canRender(p fractal.Params, w int) bool {
	return p.Unit(w) >= gpuMinUnit && p.MaxIter <= gpuMaxIter && fractal.TypeIndex(p.Type) >= 0
}

func (g *gpu) setPalette(pal *palette.Palette) {
	if pal == g.palSrc {
		return
	}
	g.palSrc = pal
	pix := make([]byte, palette.Size*4)
	for i, c := range pal.Colors {
		pix[i*4], pix[i*4+1], pix[i*4+2], pix[i*4+3] = c.R, c.G, c.B, 255
	}
	g.pal.WritePixels(pix)
}

// draw fills dst with the fractal described by p.
func (g *gpu) draw(dst *ebiten.Image, p fractal.Params, pal *palette.Palette, offset int, density float64) {
	g.setPalette(pal)
	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()
	fw, fh := float32(w), float32(h)
	g.vertices[0].DstX, g.vertices[0].DstY = 0, 0
	g.vertices[1].DstX, g.vertices[1].DstY = fw, 0
	g.vertices[2].DstX, g.vertices[2].DstY = 0, fh
	g.vertices[3].DstX, g.vertices[3].DstY = fw, fh

	cx, cy := p.CenterF()
	julia := 0.0
	if p.Julia {
		julia = 1
	}
	op := &ebiten.DrawTrianglesShaderOptions{}
	op.Images[0] = g.pal
	op.Uniforms = map[string]any{
		"Center":    []float32{float32(cx), float32(cy)},
		"Unit":      float32(p.Unit(w)),
		"Size":      []float32{fw, fh},
		"MaxIter":   float32(p.MaxIter),
		"Type":      float32(fractal.TypeIndex(p.Type)),
		"JuliaMode": float32(julia),
		"C":         []float32{float32(p.JuliaRe), float32(p.JuliaIm)},
		"Offset":    float32(offset),
		"Density":   float32(density),
	}
	dst.DrawTrianglesShader(g.vertices, g.indices, g.shader, op)
}
