// Command gofract is a cross-platform, Fractint-style fractal viewer.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/richardwooding/gofract/internal/fractal"
	"github.com/richardwooding/gofract/internal/palette"
	"github.com/richardwooding/gofract/internal/ui"
)

func main() {
	var (
		width   = flag.Int("width", 1024, "initial window width")
		height  = flag.Int("height", 768, "initial window height")
		workers = flag.Int("workers", 0, "render goroutines (0 = number of CPUs)")
		load    = flag.String("load", "", "JSON sidecar written by S to reopen a view")
		mapFile = flag.String("map", "", "Fractint .map palette to start with")
		saveDir = flag.String("out", "", "directory for saved images (default: current directory)")
		typ     = flag.String("type", "", "fractal type to start with (see -list)")
		list    = flag.Bool("list", false, "print the fractal types and exit")
		iter    = flag.Int("iter", 0, "max iterations (default 256)")
	)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: gofract [flags]\n\nPress F1 in the viewer for the key reference.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *list {
		for _, n := range fractal.Names() {
			fmt.Println(n)
		}
		return
	}

	cfg := ui.Config{
		Params:  fractal.Default(),
		Workers: *workers,
		SaveDir: *saveDir,
		MapDirs: mapDirs(),
	}
	if *load != "" {
		sc, err := ui.LoadSidecar(*load)
		if err != nil {
			log.Fatal(err)
		}
		cfg.Params = sc.Params
		cfg.Density = sc.Density
	}
	if *typ != "" {
		p, err := fractal.Start(*typ)
		if err != nil {
			log.Fatalf("%v (choose from %v)", err, fractal.Names())
		}
		cfg.Params = p
	}
	if *iter > 0 {
		cfg.Params.MaxIter = *iter
	}
	if *mapFile != "" {
		pal, err := palette.Load(*mapFile)
		if err != nil {
			log.Fatal(err)
		}
		cfg.Palette = pal
	}

	app, err := ui.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowTitle("gofract")
	ebiten.SetWindowSize(*width, *height)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}

// mapDirs lists where L looks for .map palettes: ./maps and the per-user
// config directory.
func mapDirs() []string {
	dirs := []string{"maps"}
	if c, err := os.UserConfigDir(); err == nil {
		dirs = append(dirs, filepath.Join(c, "gofract", "maps"))
	}
	return dirs
}
