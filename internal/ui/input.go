package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/richardwooding/gofract/internal/fractal"
	"github.com/richardwooding/gofract/internal/palette"
)

const (
	panFraction = 0.05
	wheelZoom   = 0.8
	keyZoom     = 0.5
	minIter     = 16
	maxIter     = 1 << 20
)

func (a *App) handleInput() error {
	switch a.mode {
	case modeHelp:
		if anyKeyJustPressed() || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			a.mode = modeView
		}
		return nil
	case modeTypeMenu:
		a.handleMenu(len(fractal.Names()), a.pickType)
		return nil
	case modeMapMenu:
		a.handleMenu(len(a.maps), a.pickMap)
		return nil
	}
	return a.handleView()
}

func anyKeyJustPressed() bool {
	return len(inpututil.AppendJustPressedKeys(nil)) > 0
}

// repeating reports a key press with keyboard-style auto repeat.
func repeating(k ebiten.Key) bool {
	d := inpututil.KeyPressDuration(k)
	return d == 1 || (d > 18 && d%3 == 0)
}

func (a *App) handleView() error {
	w, h := a.width, a.height
	p := a.params

	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
		if a.zoom.active {
			a.zoom.active = false
			return nil
		}
		return ebiten.Termination
	case inpututil.IsKeyJustPressed(ebiten.KeyF1), inpututil.IsKeyJustPressed(ebiten.KeyH):
		a.mode = modeHelp
	case inpututil.IsKeyJustPressed(ebiten.KeyT):
		a.mode = modeTypeMenu
		a.menuSel = indexOf(fractal.Names(), p.Type)
	case inpututil.IsKeyJustPressed(ebiten.KeyL):
		a.maps = findMaps(a.cfg.MapDirs)
		if len(a.maps) == 0 {
			a.flash("no .map files found in %v", a.cfg.MapDirs)
		} else {
			a.mode = modeMapMenu
			a.menuSel = 0
		}
	case inpututil.IsKeyJustPressed(ebiten.KeySpace):
		if p.Type == "julia" {
			a.setParams(p.ToMandelbrot())
		} else {
			cx, cy := ebiten.CursorPosition()
			x, y := p.PixelToPlane(float64(cx), float64(cy), w, h)
			a.setParams(p.ToJulia(x, y))
		}
	case inpututil.IsKeyJustPressed(ebiten.KeyBackspace):
		a.undo()
	case inpututil.IsKeyJustPressed(ebiten.KeyS):
		if path, err := a.save(); err != nil {
			a.flash("save failed: %v", err)
		} else {
			a.flash("saved %s", path)
		}
	case inpututil.IsKeyJustPressed(ebiten.KeyC):
		a.cycling = !a.cycling
		a.recolor = true
	case inpututil.IsKeyJustPressed(ebiten.KeyEqual), inpututil.IsKeyJustPressed(ebiten.KeyKPAdd):
		a.cycleSpeed += 0.5
		a.cycling = true
		a.flash("cycle speed %+.1f", a.cycleSpeed)
	case inpututil.IsKeyJustPressed(ebiten.KeyMinus), inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract):
		a.cycleSpeed -= 0.5
		a.cycling = true
		a.flash("cycle speed %+.1f", a.cycleSpeed)
	case inpututil.IsKeyJustPressed(ebiten.KeyP):
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			a.palIdx = (a.palIdx + len(a.palettes) - 1) % len(a.palettes)
		} else {
			a.palIdx = (a.palIdx + 1) % len(a.palettes)
		}
		a.recolor = true
		a.flash("palette %s", a.palette().Name)
	case inpututil.IsKeyJustPressed(ebiten.KeyComma):
		if p.MaxIter/2 >= minIter {
			p.MaxIter /= 2
			a.setParams(p)
		}
	case inpututil.IsKeyJustPressed(ebiten.KeyPeriod):
		if p.MaxIter*2 <= maxIter {
			p.MaxIter *= 2
			a.setParams(p)
		}
	case inpututil.IsKeyJustPressed(ebiten.KeyPageUp):
		a.setParams(p.Zoom(keyZoom))
	case inpututil.IsKeyJustPressed(ebiten.KeyPageDown):
		a.setParams(p.Zoom(1 / keyZoom))
	case inpututil.IsKeyJustPressed(ebiten.KeyHome):
		a.setParams(fractal.Default())
	case inpututil.IsKeyJustPressed(ebiten.KeyTab):
		a.showStatus = !a.showStatus
	case inpututil.IsKeyJustPressed(ebiten.KeyF):
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	case repeating(ebiten.KeyArrowLeft):
		a.setParams(p.Pan(-panFraction, 0, w, h))
	case repeating(ebiten.KeyArrowRight):
		a.setParams(p.Pan(panFraction, 0, w, h))
	case repeating(ebiten.KeyArrowUp):
		a.setParams(p.Pan(0, -panFraction, w, h))
	case repeating(ebiten.KeyArrowDown):
		a.setParams(p.Pan(0, panFraction, w, h))
	}

	// Mouse: wheel zooms about the cursor, left drag draws a zoom box,
	// right click zooms out about the cursor.
	cx, cy := ebiten.CursorPosition()
	if _, dy := ebiten.Wheel(); dy != 0 {
		f := wheelZoom
		if dy < 0 {
			f = 1 / wheelZoom
		}
		a.setParams(a.params.ZoomAt(f, float64(cx), float64(cy), w, h))
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		a.setParams(a.params.ZoomAt(2, float64(cx), float64(cy), w, h))
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		a.zoom.begin(cx, cy)
	}
	if a.zoom.active {
		a.zoom.move(cx, cy)
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			if np, ok := a.zoom.commit(a.params, w, h); ok {
				a.setParams(np)
			}
			a.zoom.active = false
		}
	}
	return nil
}

func (a *App) handleMenu(n int, pick func(int)) {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
		a.mode = modeView
	case inpututil.IsKeyJustPressed(ebiten.KeyEnter), inpututil.IsKeyJustPressed(ebiten.KeyKPEnter):
		a.mode = modeView
		if n > 0 {
			pick(a.menuSel)
		}
	case repeating(ebiten.KeyArrowUp):
		a.menuSel = (a.menuSel + n - 1) % max(n, 1)
	case repeating(ebiten.KeyArrowDown):
		a.menuSel = (a.menuSel + 1) % max(n, 1)
	default:
		for i := 0; i < n && i < 9; i++ {
			if inpututil.IsKeyJustPressed(ebiten.KeyDigit1 + ebiten.Key(i)) {
				a.mode = modeView
				pick(i)
				return
			}
		}
	}
}

func (a *App) pickType(i int) {
	name := fractal.Names()[i]
	p := a.params
	switch {
	case name == p.Type:
		return
	case name == "julia":
		p = p.ToJulia(p.JuliaRe, p.JuliaIm)
	case name == "mandelbrot":
		p = p.ToMandelbrot()
	default:
		p.Type = name
	}
	a.setParams(p)
}

func (a *App) pickMap(i int) {
	pal, err := palette.Load(a.maps[i])
	if err != nil {
		a.flash("%v", err)
		return
	}
	a.palettes = append(a.palettes, pal)
	a.palIdx = len(a.palettes) - 1
	a.recolor = true
	a.flash("palette %s", pal.Name)
}

func indexOf(ss []string, s string) int {
	for i, v := range ss {
		if v == s {
			return i
		}
	}
	return 0
}
