#!/usr/bin/env bash
# Import freshly built Flatpak bundles into the self-hosted OSTree repository
# served from GitHub Pages, sign it, and write the .flatpakrepo/.flatpakref
# files users install from.
# Usage: packaging/flatpak/publish-repo.sh <site-dir> <bundle>...
# Env: GPG_KEY_ID (signing key already imported into gpg), SITE_URL
set -euo pipefail

site=$1; shift
: "${GPG_KEY_ID:?set GPG_KEY_ID}"
: "${SITE_URL:?set SITE_URL}"   # e.g. https://richardwooding.github.io/gofract
app=io.github.richardwooding.gofract
repo="$site/repo"

if [ ! -f "$repo/config" ]; then
  ostree init --repo="$repo" --mode=archive
fi

for bundle in "$@"; do
  echo "importing $bundle"
  flatpak build-import-bundle --gpg-sign="$GPG_KEY_ID" "$repo" "$bundle"
done

flatpak build-update-repo \
  --title="gofract" \
  --comment="Fractint-style fractal viewer" \
  --description="Explore the Mandelbrot set and friends, Fractint style." \
  --homepage="https://github.com/richardwooding/gofract" \
  --icon="https://raw.githubusercontent.com/richardwooding/gofract/main/assets/icon-128.png" \
  --default-branch=master \
  --gpg-sign="$GPG_KEY_ID" \
  --prune --prune-depth=3 \
  "$repo"

key=$(base64 -w0 packaging/flatpak/gofract.gpg)
cp packaging/flatpak/gofract.gpg "$site/gofract.gpg"

cat > "$site/gofract.flatpakrepo" <<REPO
[Flatpak Repo]
Title=gofract
Url=$SITE_URL/repo/
Homepage=https://github.com/richardwooding/gofract
Comment=Fractint-style fractal viewer
Description=Explore the Mandelbrot set and friends, Fractint style.
Icon=https://raw.githubusercontent.com/richardwooding/gofract/main/assets/icon-128.png
GPGKey=$key
REPO

cat > "$site/gofract.flatpakref" <<REF
[Flatpak Ref]
Name=$app
Branch=master
Title=gofract
Url=$SITE_URL/repo/
Homepage=https://github.com/richardwooding/gofract
Icon=https://raw.githubusercontent.com/richardwooding/gofract/main/assets/icon-128.png
RuntimeRepo=https://flathub.org/repo/flathub.flatpakrepo
IsRuntime=false
GPGKey=$key
REF

cat > "$site/index.html" <<HTML
<!doctype html>
<meta charset="utf-8">
<title>gofract Flatpak repository</title>
<style>body{font-family:sans-serif;max-width:40em;margin:3em auto;padding:0 1em}code,pre{background:#eee;padding:.2em .4em}</style>
<h1>gofract Flatpak repository</h1>
<p>A Fractint-style fractal viewer. Source, other packages and documentation:
<a href="https://github.com/richardwooding/gofract">github.com/richardwooding/gofract</a>.</p>
<h2>Install</h2>
<pre>flatpak install --from $SITE_URL/gofract.flatpakref</pre>
<p>That adds this repository as a remote and installs the app; <code>flatpak update</code> picks up new releases.
The runtime comes from Flathub, which must be configured as a remote.</p>
<p>Repository signing key: <a href="gofract.gpg">gofract.gpg</a>, fingerprint <code>$GPG_KEY_ID</code>.</p>
<p>Published by the release workflow for ${GITHUB_REF_NAME:-a local build}.</p>
HTML

touch "$site/.nojekyll"
echo "published to $site"
