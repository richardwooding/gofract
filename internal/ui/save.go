package ui

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/richardwooding/gofract/internal/fractal"
)

// Sidecar is the JSON written next to each saved PNG so a view can be
// reopened with -load.
type Sidecar struct {
	Params  fractal.Params `json:"params"`
	Palette string         `json:"palette"`
	Density float64        `json:"density,omitempty"`
	Width   int            `json:"width"`
	Height  int            `json:"height"`
	Saved   time.Time      `json:"saved"`
}

// savePNG writes the current frame to path.
func (a *App) savePNG(path string) error {
	if a.img == nil {
		return fmt.Errorf("nothing rendered yet")
	}
	w, h := a.img.Bounds().Dx(), a.img.Bounds().Dy()
	if len(a.pix) != w*h*4 {
		return fmt.Errorf("pixel buffer out of sync")
	}
	if a.gpuActive {
		a.img.ReadPixels(a.pix)
	}
	img := &image.RGBA{Pix: a.pix, Stride: w * 4, Rect: image.Rect(0, 0, w, h)}
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
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

// save writes the current frame as PNG plus a JSON sidecar and returns the
// PNG path.
func (a *App) save() (string, error) {
	dir := a.cfg.SaveDir
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	base := filepath.Join(dir, "gofract-"+time.Now().Format("20060102-150405"))
	if err := a.savePNG(base + ".png"); err != nil {
		return "", err
	}
	w, h := a.img.Bounds().Dx(), a.img.Bounds().Dy()

	sc := Sidecar{Params: a.params, Palette: a.palette().Name, Density: a.density, Width: w, Height: h, Saved: time.Now()}
	data, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(base+".json", data, 0o644); err != nil {
		return "", err
	}
	return base + ".png", nil
}

// LoadSidecar reads a JSON file written by save.
func LoadSidecar(path string) (Sidecar, error) {
	var sc Sidecar
	data, err := os.ReadFile(path)
	if err != nil {
		return sc, err
	}
	if err := json.Unmarshal(data, &sc); err != nil {
		return sc, fmt.Errorf("%s: %w", path, err)
	}
	if sc.Params.Scale <= 0 || sc.Params.MaxIter <= 0 {
		return sc, fmt.Errorf("%s: missing or invalid params", path)
	}
	return sc, nil
}

// findMaps lists *.map files in the given directories, sorted by name.
func findMaps(dirs []string) []string {
	var out []string
	for _, d := range dirs {
		matches, _ := filepath.Glob(filepath.Join(d, "*.map"))
		out = append(out, matches...)
	}
	sort.Strings(out)
	return out
}
