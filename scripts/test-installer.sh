#!/bin/sh
# Exercise installation without network access or changes to shell configuration.
set -eu
workdir=$(mktemp -d)
trap 'rm -rf "$workdir"' EXIT HUP INT TERM
mkdir -p "$workdir/mock" "$workdir/assets" "$workdir/bin"
cp dist/pass-gen_* dist/checksums.txt dist/VERSION "$workdir/assets/"
cat > "$workdir/mock/curl" <<'MOCK'
#!/bin/sh
set -eu
while [ "$#" -gt 0 ]; do
  case "$1" in
    https://*) url=$1; shift ;;
    -o) output=$2; shift 2 ;;
    --proto|--proto-redir) shift 2 ;;
    *) shift ;;
  esac
done
cp "$PASS_GEN_TEST_ASSETS/${url##*/}" "$output"
MOCK
chmod +x "$workdir/mock/curl"
export PASS_GEN_TEST_ASSETS="$workdir/assets"
export PASS_GEN_INSTALL_DIR="$workdir/bin"
export PASS_GEN_NO_PATH=1
export PATH="$workdir/mock:$PATH"
sh install.sh
[ "$("$workdir/bin/pass-gen" --version)" = "pass-gen $(cat dist/VERSION)" ]
[ "$("$workdir/bin/pass-gen" --length 24 --no-copy | tr -d '\n' | wc -c | tr -d ' ')" = 24 ]
# Reject a mismatching checksum and retain the previously installed executable.
sed 's/^[0-9a-f]*/0000000000000000000000000000000000000000000000000000000000000000/' dist/checksums.txt > "$workdir/assets/checksums.txt"
if sh install.sh > "$workdir/failure.log" 2>&1; then
  echo 'ERROR: installer accepted a bad checksum' >&2
  exit 1
fi
grep -q 'Checksum mismatch' "$workdir/failure.log"
[ "$("$workdir/bin/pass-gen" --version)" = "pass-gen $(cat dist/VERSION)" ]
printf 'Installer checks passed.\n'
