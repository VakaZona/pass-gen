#!/bin/sh
set -eu
version=${1:?Usage: sh scripts/build-release.sh v0.1.1}
case "$version" in v[0-9]*.[0-9]*.[0-9]*) ;; *) echo 'Expected a version like v0.1.1' >&2; exit 1 ;; esac
case "$version" in *[!v0-9.]* ) echo 'Only stable numeric versions are supported' >&2; exit 1 ;; esac
mkdir -p dist
for platform in darwin linux; do
  for arch in amd64 arm64; do
    asset="pass-gen_${platform}_${arch}"
    CGO_ENABLED=0 GOOS="$platform" GOARCH="$arch" go build -trimpath -ldflags "-s -w -X main.version=$version" -o "dist/$asset" .
  done
done
printf '%s\n' "$version" > dist/VERSION
cp install.sh dist/install.sh
(
  cd dist
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum pass-gen_darwin_amd64 pass-gen_darwin_arm64 pass-gen_linux_amd64 pass-gen_linux_arm64 > checksums.txt
  else
    shasum -a 256 pass-gen_darwin_amd64 pass-gen_darwin_arm64 pass-gen_linux_amd64 pass-gen_linux_arm64 > checksums.txt
  fi
)
