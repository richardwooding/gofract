# gofract

A cross-platform fractal viewer in the spirit of the classic Fractint,
written in Go on [Ebitengine](https://ebitengine.org).

Hotkey driven, GPU shader rendering for shallow zooms with an automatic
hand-off to a progressive CPU renderer when float32 runs out, perturbation
deep zoom past float64 limits, palette cycling,
rubber-band zoom box, Fractint `.map` palettes. Families: Mandelbrot, Burning
Ship, Tricorn, Multibrot (z^3), Newton (z^3 - 1), each escape-time family
with its Julia variant.

## Install

Prebuilt binaries for Linux (amd64, arm64), macOS (universal) and Windows
(amd64, arm64) are attached to each
[GitHub release](https://github.com/richardwooding/gofract/releases). Unpack
and run `gofract`; nothing else is needed. macOS may ask you to allow the
unsigned binary in System Settings > Privacy & Security the first time.

To cut a release, push a tag: `git tag v0.2.0 && git push origin v0.2.0`. The
release workflow builds every platform and attaches the archives with a
checksum file.

## Build from source

Requires Go 1.27 and, on Linux and macOS, a C compiler because Ebitengine
uses cgo there. Windows needs no C toolchain.

Linux (Fedora):

    sudo dnf install gcc pkg-config mesa-libGL-devel mesa-libGLES-devel \
        libX11-devel libXcursor-devel libXi-devel libXinerama-devel \
        libXrandr-devel libXxf86vm-devel libxkbcommon-devel wayland-devel \
        alsa-lib-devel

Linux (Debian/Ubuntu):

    sudo apt install gcc pkg-config libgl1-mesa-dev libgles2-mesa-dev \
        libx11-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev \
        libxxf86vm-dev libxkbcommon-dev libwayland-dev libasound2-dev

macOS: `xcode-select --install`.

Then:

    go build -o gofract ./cmd/gofract
    ./gofract

On an immutable Fedora host (Silverblue, Bluefin) build inside a distrobox:

    distrobox create -n gofract -i registry.fedoraproject.org/fedora-toolbox:44
    distrobox enter gofract -- sudo dnf install -y golang <packages above>
    distrobox enter gofract -- go build -o gofract ./cmd/gofract

The binary runs on the host.

## Use

    gofract [-width 1024] [-height 768] [-type burningship] [-iter 1024]
            [-gpu auto|off|on] [-shot frame.png] [-list]
            [-map palette.map] [-load view.json] [-out dir] [-workers n]

Press `F1` in the viewer for the full key list. The essentials:

| Input | Action |
|---|---|
| drag left mouse | zoom box, release to zoom in |
| right click / wheel | zoom out / in about the cursor |
| arrows, PgUp, PgDn, Home | pan, zoom, reset |
| Backspace | undo the last view change |
| Space | toggle Julia mode with c at the cursor point |
| `[` `]` | halve / double colour density |
| T | choose the fractal type |
| `,` `.` | halve / double max iterations |
| C, `-`, `=` | toggle palette cycling, slower, faster |
| P | next palette |
| G | GPU shader: auto, off, on |
| L | load a `.map` from `./maps` or `$XDG_CONFIG_HOME/gofract/maps` |
| S | save PNG and a JSON sidecar reopenable with `-load` |
| Esc | quit |

`-shot file.png` renders one frame and exits, handy for scripting or for
comparing the GPU and CPU paths.

## Deep zoom

The view centre is stored as decimal text at whatever precision the zoom
needs, so panning and zooming never lose digits. Below a view width of 1e-8
the escape-time families (Mandelbrot, Burning Ship, Tricorn, Multibrot, and
their Julia sets) switch to perturbation rendering: a reference orbit at the
centre in arbitrary precision, each pixel iterating its float64 offset from
it with a family-specific recurrence, and rebasing onto the critical orbit
when a pixel drifts from the reference. Deep views need many more
iterations; press `.` until the detail resolves. Newton falls back to plain
float64 and pixelates past that depth, and the status line says so.

## Layout

    cmd/gofract        entry point and flags
    internal/fractal   Params (view maths) and the escape-time kernels
    internal/render    progressive tiled renderer on a goroutine pool
    internal/palette   256-colour palettes, presets, .map load/save
    internal/ui        Ebitengine game loop, input, overlays, saving,
                       and the Kage shader (fractal.kage)

## Tests

    go test ./...
    go test -bench . -run xxx ./internal/render/
