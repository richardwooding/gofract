# Flathub submission

This directory holds the manifest and generated module list that go into the
Flathub repository. It differs from `../flatpak/`, which builds the sideload
bundle in CI from the working tree.

## Files

- `io.github.richardwooding.gofract.yml`: the manifest. Sources are a pinned
  git tag of this repository plus the Go modules in `go.mod.yml`.
- `go.mod.yml`: Go module archives from the module proxy with checksums.
- `modules.txt`: the vendor manifest Go expects when building with
  `-mod=vendor`.

## Updating for a release

1. Tag and publish the release as usual.
2. Point the git source at the tag and its commit:
   `git rev-parse vX.Y.Z^{commit}`.
3. If `go.mod` or `go.sum` changed since the last submission, regenerate:

       go install github.com/dennwc/flatpak-go-mod@latest
       cd packaging/flathub && flatpak-go-mod ../..

   CI fails if these files are stale, so this step is hard to forget.
4. Push; the `flathub` CI job lints the manifest and metainfo with
   `flatpak-builder-lint`, builds the app exactly as Flathub will, and lints
   the resulting repository.
5. Copy the three files into your checkout of the Flathub repository and open
   a pull request there.

## First submission

1. Fork https://github.com/flathub/flathub and create a branch from the
   `new-pr` branch (not `master`).
2. Add the three files from this directory at the root of that branch.
3. Open a pull request against `flathub:new-pr`. A bot test-builds it;
   reviewers may ask about permissions or metadata.
4. After the merge Flathub creates `flathub/io.github.richardwooding.gofract`
   and adds you as its maintainer. Builds publish from there, and the
   `x-checker-data` block lets Flathub's external data checker open update
   pull requests automatically when a new tag appears.
5. Sign in to https://flathub.org with GitHub to verify ownership of the
   `io.github.richardwooding.*` ID.
