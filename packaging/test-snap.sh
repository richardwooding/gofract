#!/usr/bin/env bash
# Install a freshly built snap with snapd and prove it launches under strict
# confinement: CLI flags, then a headless render through Xvfb so the window,
# GL (software via the gnome extension's Mesa) and the home plug all work.
# Usage: packaging/test-snap.sh path/to/gofract.snap
set -euo pipefail

snap_file=$1
sudo snap install --dangerous "$snap_file"
snap list gofract

bin=/snap/bin/gofract
"$bin" -version
# Capture to a file first: an early-exiting grep would send SIGPIPE to
# gofract and fail the pipeline under pipefail.
"$bin" -list > "$HOME/gofract-list.txt"
grep -qx mandelbrot "$HOME/gofract-list.txt"

if ! command -v xvfb-run >/dev/null; then
  sudo apt-get update -qq
  sudo apt-get install -y -qq xvfb
fi

# The home plug grants access to visible files in the real home directory.
out="$HOME/gofract-snap-test.png"
rm -f "$out"
xvfb-run -a -s "-screen 0 640x480x24" "$bin" -width 320 -height 240 -shot "$out"

# A PNG header and a plausible size: a 320x240 fractal is tens of kilobytes.
head -c 8 "$out" | od -An -c | tr -d ' \n' | grep -q '211PNG' || { echo "not a PNG"; exit 1; }
size=$(stat -c %s "$out")
echo "rendered $out ($size bytes)"
[ "$size" -gt 2000 ] || { echo "image suspiciously small"; exit 1; }

sudo snap remove gofract
