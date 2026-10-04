package ui

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// The Kage source is checked at compile time of the shader, not of the
// program, so make sure it still parses.
func TestShaderCompiles(t *testing.T) {
	if _, err := ebiten.NewShader(kageSrc); err != nil {
		t.Fatal(err)
	}
}
