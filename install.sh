#!/bin/sh
set -eu
repo=https://github.com/VakaZona/pass-gen
fail() { printf 'pass-gen: %s\n' "$*" >&2; exit 1; }
command -v curl >/dev/null 2>&1 || fail 'Install curl first.'
case "$(uname -s)" in Darwin) platform=darwin ;; Linux) platform=linux ;; *) fail 'Supported systems: macOS and Linux.' ;; esac
case "$(uname -m)" in arm64|aarch64) arch=arm64 ;; x86_64|amd64) arch=amd64 ;; *) fail 'Supported architectures: x86_64 and ARM64.' ;; esac
workdir=$(mktemp -d)
trap 'rm -rf "$workdir"' EXIT HUP INT TERM
fetch() { curl --fail --silent --show-error --location --proto '=https' --proto-redir '=https' "$1" -o "$2"; }
version=${PASS_GEN_VERSION:-}
if [ -z "$version" ]; then
  fetch "$repo/releases/latest/download/VERSION" "$workdir/VERSION"
  version=$(cat "$workdir/VERSION")
fi
case "$version" in v[0-9]*.[0-9]*.[0-9]*) ;; *) fail 'Invalid release version.' ;; esac
case "$version" in *[!v0-9.]* ) fail 'Invalid release version.' ;; esac
asset="pass-gen_${platform}_${arch}"
base="$repo/releases/download/$version"
printf 'Installing pass-gen %s (%s/%s)...\n' "$version" "$platform" "$arch"
fetch "$base/$asset" "$workdir/$asset"
fetch "$base/checksums.txt" "$workdir/checksums.txt"
expected=$(awk -v file="$asset" '$2 == file {print $1}' "$workdir/checksums.txt")
[ "${#expected}" -eq 64 ] || fail 'Missing or invalid SHA-256 checksum.'
case "$expected" in *[!0-9a-f]*) fail 'Invalid SHA-256 checksum.' ;; esac
if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$workdir/$asset" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
  actual=$(shasum -a 256 "$workdir/$asset" | awk '{print $1}')
else
  fail 'Install sha256sum or shasum first.'
fi
[ "$actual" = "$expected" ] || fail 'Checksum mismatch; installation cancelled.'
install_dir=${PASS_GEN_INSTALL_DIR:-"$HOME/.local/bin"}
mkdir -p "$install_dir"
# Replace the executable atomically, retaining settings outside this directory.
staged=$(mktemp "$install_dir/.pass-gen.XXXXXX")
trap 'rm -rf "$workdir"; rm -f "$staged"' EXIT HUP INT TERM
cp "$workdir/$asset" "$staged"
chmod 755 "$staged"
mv -f "$staged" "$install_dir/pass-gen"
"$install_dir/pass-gen" --version
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *)
    if [ "$install_dir" = "$HOME/.local/bin" ] && [ "${PASS_GEN_NO_PATH:-0}" != 1 ]; then
      case "${SHELL:-}" in
        */zsh) profile="${ZDOTDIR:-$HOME}/.zshrc" ;;
        */bash)
          if [ "$platform" = darwin ]; then profile="$HOME/.bash_profile"; else profile="$HOME/.bashrc"; fi ;;
        *) profile= ;;
      esac
      if [ -n "$profile" ]; then
        line='export PATH="$HOME/.local/bin:$PATH" # pass-gen'
        if ! test -f "$profile" || ! grep -Fqx "$line" "$profile"; then
          printf '\n%s\n' "$line" >> "$profile"
        fi
        printf 'PATH updated in %s. Open a new terminal, then run pass-gen.\n' "$profile"
      else
        printf 'Add %s to PATH in your shell settings.\n' "$install_dir"
      fi
    else
      printf 'Add %s to PATH.\n' "$install_dir"
    fi
    ;;
esac
printf 'Installed: %s/pass-gen\n' "$install_dir"
