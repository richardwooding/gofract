# gofract

A cross-platform fractal viewer in the spirit of the classic Fractint,
written in Go on [Ebitengine](https://ebitengine.org).

Hotkey driven, progressive rendering from coarse to fine, palette cycling,
rubber-band zoom box, Mandelbrot and Julia sets, Fractint `.map` palettes.

## Build

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

    gofract [-width 1024] [-height 768] [-type julia] [-iter 1024]
            [-map palette.map] [-load view.json] [-out dir] [-workers n]

Press `F1` in the viewer for the full key list. The essentials:

| Input | Action |
|---|---|
| drag left mouse | zoom box, release to zoom in |
| right click / wheel | zoom out / in about the cursor |
| arrows, PgUp, PgDn, Home | pan, zoom, reset |
| Backspace | undo the last view change |
| Space | toggle Mandelbrot and Julia at the cursor point |
| T | choose the fractal type |
| `,` `.` | halve / double max iterations |
| C, `-`, `=` | toggle palette cycling, slower, faster |
| P | next palette |
| L | load a `.map` from `./maps` or `$XDG_CONFIG_HOME/gofract/maps` |
| S | save PNG and a JSON sidecar reopenable with `-load` |
| Esc | quit |

## Layout

    cmd/gofract        entry point and flags
    internal/fractal   Params (view maths) and the escape-time kernels
    internal/render    progressive tiled renderer on a goroutine pool
    internal/palette   256-colour palettes, presets, .map load/save
    internal/ui        Ebitengine game loop, input, overlays, saving

## Tests

    go test ./...
    go test -bench . -run xxx ./internal/render/
