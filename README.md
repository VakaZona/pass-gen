# pass-gen

**A password in your terminal. Ready to paste.**

[![Release](https://img.shields.io/github/v/release/VakaZona/pass-gen)](https://github.com/VakaZona/pass-gen/releases)
[![Release workflow](https://github.com/VakaZona/pass-gen/actions/workflows/release.yml/badge.svg)](https://github.com/VakaZona/pass-gen/actions/workflows/release.yml)
![Platforms](https://img.shields.io/badge/platforms-macOS%20%7C%20Linux-blue)

Generate a random password, print it, and copy it to your clipboard with one command. Save your preferences once and use them on every run.

- **Secure randomness** using Go's `crypto/rand`.
- **Your defaults** for length, letters, digits, and symbols.
- **Ready to paste** with automatic clipboard copying.
- **Standalone binaries** for macOS and Linux on x86_64 and ARM64.
- **No password storage** — only generation settings are saved.

## Install

```sh
curl -fsSL https://github.com/VakaZona/pass-gen/releases/latest/download/install.sh | sh
```

Open a new terminal, then run:

```sh
pass-gen
```

The installer downloads the latest release for your system, verifies its SHA-256 checksum, and installs it to `~/.local/bin/pass-gen`. It adds that directory to your zsh or bash configuration if needed. Other shells require a manual PATH update.

No Go, Git, Homebrew, or `sudo` required. You need `curl` and either `shasum` or `sha256sum`. Release downloads become available after the first successful release workflow.

> **Updating:** run the same install command again. Your saved settings are preserved.

<details>
<summary>Installation options</summary>

To review the installer before running it:

```sh
curl -fL https://github.com/VakaZona/pass-gen/releases/latest/download/install.sh -o install.sh
# Read install.sh, then:
sh install.sh
```

To install a specific published version, replace `v0.1.1` as needed:

```sh
curl -fsSL https://github.com/VakaZona/pass-gen/releases/download/v0.1.1/install.sh \
  | PASS_GEN_VERSION=v0.1.1 sh
```

| Environment variable | Purpose |
| --- | --- |
| `PASS_GEN_VERSION` | Install a specific release, such as `v0.1.1`. |
| `PASS_GEN_INSTALL_DIR` | Override the install directory. Add it to PATH yourself. |
| `PASS_GEN_NO_PATH=1` | Leave your shell configuration unchanged. |

You can also download a binary directly from [Releases](https://github.com/VakaZona/pass-gen/releases).

</details>

### Clipboard on Linux

macOS uses its built-in `pbcopy`. On Linux, install a clipboard tool for your desktop session:

| Session | Required tool | Ubuntu / Debian |
| --- | --- | --- |
| Wayland | `wl-copy` | `sudo apt install wl-clipboard` |
| X11 | `xclip` or `xsel` | `sudo apt install xclip` |

On a server or over SSH without a desktop clipboard, use `pass-gen --no-copy`.

## Usage

```sh
pass-gen                            # Generate, print, and copy
pass-gen --length 32                 # Override length for this run
pass-gen --symbols=false             # Skip symbols for this run
pass-gen --length 24 --symbols=false # Combine overrides
pass-gen --no-copy                   # Print without copying
pass-gen --version                   # Show installed version
pass-gen --help                      # Show all options
```

Command-line options apply only to the current run. Use `--option=false` to disable a boolean option.

## Configuration

```sh
pass-gen config                         # Show current settings
pass-gen config set length 24            # Save a new default length
pass-gen config set symbols false        # Disable symbols by default
pass-gen config set exclude_ambiguous true
pass-gen config path                    # Show the config file location
pass-gen config reset                   # Restore defaults
```

Changes take effect on the next run.

| Setting | Default | Description |
| --- | --- | --- |
| `length` | `20` | Number of characters, from 1 to 4096. |
| `lowercase` | `true` | Include lowercase Latin letters. |
| `uppercase` | `true` | Include uppercase Latin letters. |
| `digits` | `true` | Include digits. |
| `symbols` | `true` | Include special characters. |
| `exclude_ambiguous` | `true` | Exclude `i`, `l`, `I`, `o`, `O`, `0`, and `1`. |

Each password contains at least one character from every enabled category. At least one category must be enabled, and the length must be at least the number of enabled categories.

Available symbols:

```text
!@#$%^&*()-_=+[]{};:,.?/~
```

Settings are stored in `~/.config/pass-gen/config.json`. If `XDG_CONFIG_HOME` is set to an absolute path, that directory is used instead. The file is created when you save or reset settings; built-in defaults apply until then.

**Settings priority:** command-line options → saved configuration → built-in defaults.

## Output behavior

On a successful generation, stdout contains only the password followed by a newline. The clipboard receives the password without the newline.

If copying fails, the password is still printed, an error is written to stderr, and the command exits with status `1`. Use `--no-copy` when copying is unnecessary.

Passwords are not saved by the application. Printed passwords remain in terminal scrollback, and copied passwords remain in the clipboard until it changes.

## Development

Requires **Go 1.25 or later**.

```sh
git clone https://github.com/VakaZona/pass-gen.git
cd pass-gen
make build
./bin/pass-gen
```

Run the tests:

```sh
make test
```

Build all four release binaries and check the installer locally:

```sh
sh scripts/build-release.sh v0.1.1
sh scripts/test-installer.sh
```

The installer check uses local fixtures without downloading files or changing your shell configuration.

### Publish a release

Commit and push your changes, then create a new stable version tag:

```sh
git push origin main
git tag v0.1.1
git push origin v0.1.1
```

Use a new `vMAJOR.MINOR.PATCH` tag for each release. Replace `v0.1.1` with your next version; do not move existing tags.

[GitHub Actions](https://github.com/VakaZona/pass-gen/actions) runs the tests, builds the macOS and Linux binaries, checks the installer, and publishes a [release](https://github.com/VakaZona/pass-gen/releases) containing the binaries, installer, version file, and checksums. The workflow uses GitHub's built-in token; no additional secret is required.
