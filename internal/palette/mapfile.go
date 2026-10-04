package palette

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Parse reads a Fractint .map file: one "r g b" triple per line, optional
// trailing text, blank lines ignored. Fewer than 256 entries are padded by
// repeating the last colour; extra entries are ignored.
func Parse(name string, r io.Reader) (*Palette, error) {
	p := &Palette{Name: name}
	n := 0
	sc := bufio.NewScanner(r)
	line := 0
	for sc.Scan() && n < Size {
		line++
		fields := strings.Fields(sc.Text())
		if len(fields) == 0 {
			continue
		}
		if len(fields) < 3 {
			return nil, fmt.Errorf("map %s line %d: expected 3 values, got %d", name, line, len(fields))
		}
		var c [3]uint8
		for i := 0; i < 3; i++ {
			v, err := strconv.Atoi(fields[i])
			if err != nil || v < 0 || v > 255 {
				return nil, fmt.Errorf("map %s line %d: bad value %q", name, line, fields[i])
			}
			c[i] = uint8(v)
		}
		p.Colors[n] = rgb(c[0], c[1], c[2])
		n++
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, fmt.Errorf("map %s: no colours", name)
	}
	for i := n; i < Size; i++ {
		p.Colors[i] = p.Colors[n-1]
	}
	return p, nil
}

// Load reads a .map file from disk. The palette is named after the file.
func Load(path string) (*Palette, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return Parse(name, f)
}

// Write emits the palette in .map format.
func (p *Palette) Write(w io.Writer) error {
	bw := bufio.NewWriter(w)
	for _, c := range p.Colors {
		if _, err := fmt.Fprintf(bw, "%3d %3d %3d\n", c.R, c.G, c.B); err != nil {
			return err
		}
	}
	return bw.Flush()
}

// Save writes the palette to a .map file.
func (p *Palette) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := p.Write(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
