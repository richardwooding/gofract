package ui

import (
	"fmt"
	"image/color"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/richardwooding/gofract/internal/render"
)

const lineHeight = 14

var (
	colText   = color.RGBA{255, 255, 255, 255}
	colShadow = color.RGBA{0, 0, 0, 255}
	colHilite = color.RGBA{255, 255, 0, 255}
	colPanel  = color.RGBA{0, 0, 0, 190}
)

// drawText draws s at (x, y) with a one pixel drop shadow so it stays
// readable over any palette.
func (a *App) drawText(dst *ebiten.Image, s string, x, y float64, clr color.Color) {
	op := &text.DrawOptions{}
	op.LineSpacing = lineHeight
	op.GeoM.Translate(x+1, y+1)
	op.ColorScale.ScaleWithColor(colShadow)
	text.Draw(dst, s, a.face, op)
	op.GeoM.Reset()
	op.GeoM.Translate(x, y)
	op.ColorScale.Reset()
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(dst, s, a.face, op)
}

// panel draws a translucent box sized to fit s with padding and returns the
// text origin inside it.
func (a *App) panel(dst *ebiten.Image, s string, x, y float64) (tx, ty float64) {
	const pad = 8
	w, h := text.Measure(s, a.face, lineHeight)
	vector.FillRect(dst, float32(x), float32(y), float32(w+2*pad), float32(h+2*pad), colPanel, false)
	return x + pad, y + pad
}

func (a *App) drawStatus(dst *ebiten.Image) {
	p := a.params
	digits := int(math.Max(6, -math.Log10(p.Scale)+5))
	var b strings.Builder
	fmt.Fprintf(&b, "%s  centre %.*g %+.*gi  width %.4g  iter %d", p.Type, digits, p.CenterX, digits, p.CenterY, p.Scale, p.MaxIter)
	if p.Type == "julia" {
		fmt.Fprintf(&b, "  c=%.6g%+.6gi", p.JuliaRe, p.JuliaIm)
	}
	fmt.Fprintf(&b, "  pal %s", a.palette().Name)
	if a.cycling {
		fmt.Fprintf(&b, " cyc %+.1f", a.cycleSpeed)
	}
	if a.renderer.Busy() {
		fmt.Fprintf(&b, "  pass %d/%d", a.renderer.Pass()+1, render.Passes())
	}
	fmt.Fprintf(&b, "  %.0f fps", ebiten.ActualFPS())
	line := b.String()
	if a.msg != "" && time.Now().Before(a.msgUntil) {
		line += "\n" + a.msg
	}
	_, h := text.Measure(line, a.face, lineHeight)
	a.drawText(dst, line, 6, float64(a.height)-h-4, colText)
	if a.mode == modeView {
		a.drawText(dst, "F1 help", float64(a.width)-56, 4, colText)
	}
}

const helpText = `gofract - a Fractint-style fractal viewer

Mouse
  drag left        zoom box, release to zoom in
  right click      zoom out 2x about the cursor
  wheel            zoom in/out about the cursor

Keys
  arrows           pan
  PgUp / PgDn      zoom in / out 2x
  Home             reset view
  Backspace        undo last view change
  Space            Mandelbrot <-> Julia at cursor
  T                choose fractal type
  , / .            halve / double max iterations
  C                toggle palette cycling
  - / =            cycle slower / faster (sign reverses)
  P / Shift+P      next / previous palette
  L                load a Fractint .map palette
  S                save PNG + JSON of this view
  Tab              toggle status line
  F                toggle fullscreen
  F1 or H          this help
  Esc              quit

Press any key to close`

func (a *App) drawHelp(dst *ebiten.Image) {
	w, h := text.Measure(helpText, a.face, lineHeight)
	x := (float64(a.width) - w) / 2
	y := (float64(a.height) - h) / 2
	tx, ty := a.panel(dst, helpText, x-8, y-8)
	a.drawText(dst, helpText, tx, ty, colText)
}

func (a *App) drawMenu(dst *ebiten.Image, title string, items []string, sel int) {
	var lines []string
	lines = append(lines, title, "")
	for i, it := range items {
		marker := "  "
		if i == sel {
			marker = "> "
		}
		lines = append(lines, fmt.Sprintf("%s%d. %s", marker, i+1, it))
	}
	lines = append(lines, "", "Up/Down, Enter, Esc")
	body := strings.Join(lines, "\n")
	w, h := text.Measure(body, a.face, lineHeight)
	x := (float64(a.width) - w) / 2
	y := (float64(a.height) - h) / 2
	tx, ty := a.panel(dst, body, x-8, y-8)
	for i, l := range lines {
		clr := colText
		if i-2 == sel && i >= 2 && i-2 < len(items) {
			clr = colHilite
		}
		a.drawText(dst, l, tx, ty+float64(i*lineHeight), clr)
	}
}

func (a *App) mapLabels() []string {
	out := make([]string, len(a.maps))
	for i, m := range a.maps {
		out[i] = filepath.Base(m)
	}
	return out
}
