package palette

import (
	"bytes"
	"image/color"
	"strings"
	"testing"
)

func TestPresetsHaveBlackInsideAndFullRing(t *testing.T) {
	ps := Presets()
	if len(ps) == 0 {
		t.Fatal("no presets")
	}
	for _, p := range ps {
		if p.Colors[0] != (color.RGBA{0, 0, 0, 255}) {
			t.Errorf("%s: index 0 = %v, want black", p.Name, p.Colors[0])
		}
		for i, c := range p.Colors {
			if c.A != 255 {
				t.Errorf("%s[%d]: alpha %d", p.Name, i, c.A)
			}
		}
	}
}

func TestGradientEndpoints(t *testing.T) {
	p := Gradient("g", Stop{0, rgb(10, 20, 30)}, Stop{1, rgb(250, 240, 230)})
	if p.Colors[1] != rgb(10, 20, 30) {
		t.Errorf("first ring colour %v", p.Colors[1])
	}
	if p.Colors[255] != rgb(250, 240, 230) {
		t.Errorf("last ring colour %v", p.Colors[255])
	}
}

func TestMapRoundTrip(t *testing.T) {
	want := rainbow()
	var buf bytes.Buffer
	if err := want.Write(&buf); err != nil {
		t.Fatal(err)
	}
	got, err := Parse("rainbow", &buf)
	if err != nil {
		t.Fatal(err)
	}
	if got.Colors != want.Colors {
		t.Error("colours changed across write/parse")
	}
}

func TestParseTolerant(t *testing.T) {
	src := "  0 0 0\n\n255 0 0 ; red with a comment\n0 255 0 trailing words\n"
	p, err := Parse("t", strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if p.Colors[1] != rgb(255, 0, 0) || p.Colors[2] != rgb(0, 255, 0) {
		t.Errorf("parsed %v %v", p.Colors[1], p.Colors[2])
	}
	// Short maps pad with the last colour.
	if p.Colors[255] != rgb(0, 255, 0) {
		t.Errorf("padding = %v", p.Colors[255])
	}
}

func TestParseErrors(t *testing.T) {
	for _, src := range []string{"", "1 2\n", "1 2 300\n", "a b c\n"} {
		if _, err := Parse("bad", strings.NewReader(src)); err == nil {
			t.Errorf("%q: expected error", src)
		}
	}
}

func TestHSV(t *testing.T) {
	cases := map[float64]color.RGBA{0: rgb(255, 0, 0), 120: rgb(0, 255, 0), 240: rgb(0, 0, 255)}
	for h, want := range cases {
		if got := HSV(h, 1, 1); got != want {
			t.Errorf("HSV(%v) = %v, want %v", h, got, want)
		}
	}
}
