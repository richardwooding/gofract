#!/usr/bin/env bash
# Assemble gofract.app around a built binary and wrap it in a .dmg.
# Runs on macOS (needs iconutil, plutil, codesign, hdiutil).
# Usage: packaging/macos/package.sh <binary> <version-tag> <output.dmg>
set -euo pipefail

bin=$1
tag=$2
dmg=$3

# CFBundleVersion wants dotted digits; non-release builds get 0.0.0.
version=${tag#v}
[[ $version =~ ^[0-9]+(\.[0-9]+)*$ ]] || version=0.0.0

work=$(mktemp -d)
app="$work/gofract.app"
mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"
cp "$bin" "$app/Contents/MacOS/gofract"
chmod 755 "$app/Contents/MacOS/gofract"

# Icon: an iconset folder of the sizes Apple expects, converted to .icns.
iconset="$work/gofract.iconset"
mkdir -p "$iconset"
cp assets/icon-16.png   "$iconset/icon_16x16.png"
cp assets/icon-32.png   "$iconset/icon_16x16@2x.png"
cp assets/icon-32.png   "$iconset/icon_32x32.png"
cp assets/icon-64.png   "$iconset/icon_32x32@2x.png"
cp assets/icon-128.png  "$iconset/icon_128x128.png"
cp assets/icon-256.png  "$iconset/icon_128x128@2x.png"
cp assets/icon-256.png  "$iconset/icon_256x256.png"
cp assets/icon-512.png  "$iconset/icon_256x256@2x.png"
cp assets/icon-512.png  "$iconset/icon_512x512.png"
cp assets/icon-1024.png "$iconset/icon_512x512@2x.png"
iconutil -c icns "$iconset" -o "$app/Contents/Resources/gofract.icns"

sed "s/@VERSION@/$version/g" packaging/macos/Info.plist > "$app/Contents/Info.plist"
plutil -lint "$app/Contents/Info.plist"

# Ad-hoc signature: no identity, but Apple Silicon refuses to run unsigned
# code at all, and this keeps Gatekeeper's message to "unidentified
# developer" rather than "damaged".
codesign --force --deep --sign - "$app"
codesign --verify --deep --strict "$app"

# Disk image with the app and an Applications shortcut to drag it onto.
root="$work/dmgroot"
mkdir -p "$root"
cp -R "$app" "$root/"
ln -s /Applications "$root/Applications"
rm -f "$dmg"
hdiutil create -volname gofract -srcfolder "$root" -ov -format UDZO "$dmg"
echo "created $dmg"
