// Package ui is the Ebitengine front end: a Fractint-style hotkey-driven
// viewer that colours the renderer's iteration buffer through a cycling
// palette.
package ui

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/richardwooding/gofract/internal/fractal"
	"github.com/richardwooding/gofract/internal/palette"
	"github.com/richardwooding/gofract/internal/render"
)

type mode int

const (
	modeView mode = iota
	modeHelp
	modeTypeMenu
	modeMapMenu
)

// Config is what main hands the App on startup.
type Config struct {
	Params  fractal.Params
	Palette *palette.Palette // optional extra palette, selected at start
	Workers int
	Density float64  // palette steps per iteration; 0 means 1
	SaveDir string   // where S writes PNG + JSON; "" means current directory
	MapDirs []string // directories searched for .map files
}

// App implements ebiten.Game.
type App struct {
	cfg     Config
	params  fractal.Params
	fract   fractal.Fractal
	history []fractal.Params

	renderer *render.Renderer
	width    int
	height   int
	iters    []float32
	pix      []byte
	img      *ebiten.Image
	dirty    bool // params or size changed; restart the render
	recolor  bool // palette changed; rebuild pixels even if idle
	wasBusy  bool

	palettes    []*palette.Palette
	palIdx      int
	cycleOffset float64
	cycleSpeed  float64
	cycling     bool
	density     float64 // palette steps per iteration

	mode       mode
	menuSel    int
	maps       []string
	zoom       zoomBox
	face       text.Face
	showStatus bool
	msg        string
	msgUntil   time.Time
}

// New builds the App. It does not open a window; call ebiten.RunGame.
func New(cfg Config) (*App, error) {
	f, err := fractal.New(cfg.Params)
	if err != nil {
		return nil, err
	}
	a := &App{
		cfg:        cfg,
		params:     cfg.Params,
		fract:      f,
		renderer:   render.New(cfg.Workers),
		palettes:   palette.Presets(),
		cycleSpeed: 1,
		density:    cfg.Density,
		face:       text.NewGoXFace(bitmapfont.Face),
		showStatus: true,
	}
	if a.density <= 0 {
		a.density = 1
	}
	if cfg.Palette != nil {
		a.palettes = append([]*palette.Palette{cfg.Palette}, a.palettes...)
	}
	return a, nil
}

// Layout implements ebiten.Game. The logical screen is the window at device
// pixel density so every fractal pixel is a real pixel.
func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	scale := ebiten.Monitor().DeviceScaleFactor()
	w := int(float64(outsideWidth) * scale)
	h := int(float64(outsideHeight) * scale)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	if w != a.width || h != a.height {
		a.width, a.height = w, h
		a.dirty = true
	}
	return w, h
}

// Update implements ebiten.Game.
func (a *App) Update() error {
	if a.width == 0 {
		return nil
	}
	if err := a.handleInput(); err != nil {
		return err
	}
	if a.dirty {
		a.restart()
	}
	if a.cycling {
		a.cycleOffset += a.cycleSpeed
		for a.cycleOffset >= palette.Ring {
			a.cycleOffset -= palette.Ring
		}
		for a.cycleOffset < 0 {
			a.cycleOffset += palette.Ring
		}
	}
	return nil
}

func (a *App) restart() {
	a.dirty = false
	f, err := fractal.New(a.params)
	if err != nil {
		a.flash("%v", err)
		return
	}
	a.fract = f
	a.renderer.Start(f, a.params, a.width, a.height)
}

// setParams records the current view for undo and schedules a re-render.
func (a *App) setParams(p fractal.Params) {
	if p == a.params {
		return
	}
	a.history = append(a.history, a.params)
	if len(a.history) > 256 {
		a.history = a.history[1:]
	}
	a.params = p
	a.dirty = true
}

func (a *App) undo() {
	if len(a.history) == 0 {
		a.flash("nothing to undo")
		return
	}
	a.params = a.history[len(a.history)-1]
	a.history = a.history[:len(a.history)-1]
	a.dirty = true
}

func (a *App) palette() *palette.Palette { return a.palettes[a.palIdx] }

func (a *App) flash(format string, args ...any) {
	a.msg = fmt.Sprintf(format, args...)
	a.msgUntil = time.Now().Add(3 * time.Second)
}

// Draw implements ebiten.Game.
func (a *App) Draw(screen *ebiten.Image) {
	var w, h int
	a.iters, w, h = a.renderer.Snapshot(a.iters)
	if w == 0 || h == 0 {
		return
	}
	if a.img == nil || a.img.Bounds().Dx() != w || a.img.Bounds().Dy() != h {
		if a.img != nil {
			a.img.Deallocate()
		}
		a.img = ebiten.NewImage(w, h)
		a.pix = make([]byte, w*h*4)
		a.recolor = true
	}
	busy := a.renderer.Busy()
	if busy || a.wasBusy || a.cycling || a.recolor {
		a.colorize()
		a.img.WritePixels(a.pix)
		a.recolor = false
	}
	a.wasBusy = busy
	screen.DrawImage(a.img, nil)

	a.zoom.draw(screen, w, h)
	switch a.mode {
	case modeHelp:
		a.drawHelp(screen)
	case modeTypeMenu:
		a.drawMenu(screen, "Fractal type", fractal.Names(), a.menuSel)
	case modeMapMenu:
		a.drawMenu(screen, "Palette map", a.mapLabels(), a.menuSel)
	}
	if a.showStatus {
		a.drawStatus(screen)
	}
}

// colorize maps the iteration buffer through the palette into a.pix.
func (a *App) colorize() {
	pal := a.palette()
	off := int(a.cycleOffset)
	dens := float32(a.density)
	pix := a.pix
	inside := pal.Colors[0]
	for i, v := range a.iters {
		o := i * 4
		if v < 0 {
			pix[o], pix[o+1], pix[o+2], pix[o+3] = inside.R, inside.G, inside.B, 255
			continue
		}
		idx := (int(v*dens) + off) % palette.Ring
		c := pal.Colors[1+idx]
		pix[o], pix[o+1], pix[o+2], pix[o+3] = c.R, c.G, c.B, 255
	}
}
