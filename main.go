package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type Config struct {
	Length           int  `json:"length"`
	Uppercase        bool `json:"uppercase"`
	Lowercase        bool `json:"lowercase"`
	Digits           bool `json:"digits"`
	Symbols          bool `json:"symbols"`
	ExcludeAmbiguous bool `json:"exclude_ambiguous"`
}

func defaults() Config { return Config{20, true, true, true, true, true} }
func configPath() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		if !filepath.IsAbs(dir) {
			return "", errors.New("XDG_CONFIG_HOME must be an absolute path")
		}
		return filepath.Join(dir, "pass-gen", "config.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "pass-gen", "config.json"), nil
}
func loadConfig(path string) (Config, error) {
	c := defaults()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return c, fmt.Errorf("invalid config %s: %w", path, err)
	}
	return c, nil
}
func saveConfig(path string, c Config) error {
	if _, err := c.groups(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".config-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(append(data, '\n')); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
func (c Config) groups() ([]string, error) {
	if c.Length < 1 || c.Length > 4096 {
		return nil, errors.New("length must be between 1 and 4096")
	}
	var groups []string
	for _, group := range []struct {
		enabled bool
		chars   string
	}{
		{c.Lowercase, "abcdefghijklmnopqrstuvwxyz"},
		{c.Uppercase, "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{c.Digits, "0123456789"},
		{c.Symbols, "!@#$%^&*()-_=+[]{};:,.?/~"},
	} {
		if !group.enabled {
			continue
		}
		chars := group.chars
		if c.ExcludeAmbiguous {
			chars = strings.Map(func(r rune) rune {
				if strings.ContainsRune("ilIoO01", r) {
					return -1
				}
				return r
			}, chars)
		}
		groups = append(groups, chars)
	}
	if len(groups) == 0 {
		return nil, errors.New("enable at least one character category")
	}
	if c.Length < len(groups) {
		return nil, errors.New("length must be at least the number of enabled character categories")
	}
	return groups, nil
}
func randomIndex(n int) (int, error) {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0, err
	}
	return int(v.Int64()), nil
}
func generate(c Config) (string, error) {
	groups, err := c.groups()
	if err != nil {
		return "", err
	}
	alphabet := strings.Join(groups, "")
	// Rejection sampling keeps every valid password equally likely.
	for {
		password := make([]byte, c.Length)
		for i := range password {
			j, err := randomIndex(len(alphabet))
			if err != nil {
				return "", err
			}
			password[i] = alphabet[j]
		}
		valid := true
		for _, group := range groups {
			if !strings.ContainsAny(string(password), group) {
				valid = false
				break
			}
		}
		if valid {
			return string(password), nil
		}
	}
}

var version = "dev"

func copyPassword(password string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("/usr/bin/pbcopy")
	case "linux":
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			if path, err := exec.LookPath("wl-copy"); err == nil {
				cmd = exec.Command(path)
			}
		}
		if cmd == nil && os.Getenv("DISPLAY") != "" {
			if path, err := exec.LookPath("xclip"); err == nil {
				cmd = exec.Command(path, "-selection", "clipboard")
			} else if path, err := exec.LookPath("xsel"); err == nil {
				cmd = exec.Command(path, "--clipboard", "--input")
			}
		}
		if cmd == nil {
			return errors.New("Linux clipboard requires a desktop session and wl-copy (wl-clipboard), xclip or xsel; use --no-copy on a server")
		}
	default:
		return errors.New("clipboard is supported on macOS and Linux only")
	}
	cmd.Stdin = strings.NewReader(password)
	return cmd.Run()
}

const help = `pass-gen — generate a password and copy it to the clipboard

Usage:
  pass-gen [options]
  pass-gen config                 Show current settings
  pass-gen config set KEY VALUE   Save a setting
  pass-gen config reset           Restore defaults
  pass-gen config path            Show config file path

Options (override saved settings for this run):
  --length N
  --uppercase=true|false
  --lowercase=true|false
  --digits=true|false
  --symbols=true|false
  --exclude-ambiguous=true|false   Exclude i l I o O 0 1
  --no-copy                       Print without copying
  --help
  --version

Config keys: length, uppercase, lowercase, digits, symbols, exclude_ambiguous
`

func run(args []string, out io.Writer, copyFn func(string) error) error {
	if len(args) == 1 && args[0] == "--version" {
		_, err := fmt.Fprintln(out, "pass-gen", version)
		return err
	}
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		_, err := io.WriteString(out, help)
		return err
	}
	path, err := configPath()
	if err != nil {
		return err
	}
	if len(args) > 0 && args[0] == "config" {
		if len(args) == 2 && args[1] == "path" {
			_, err := fmt.Fprintln(out, path)
			return err
		}
		if len(args) == 2 && args[1] == "reset" {
			return saveConfig(path, defaults())
		}
		c, err := loadConfig(path)
		if err != nil {
			return err
		}
		if len(args) == 1 {
			enc := json.NewEncoder(out)
			enc.SetIndent("", "  ")
			return enc.Encode(c)
		}
		if len(args) != 4 || args[1] != "set" {
			return errors.New("usage: pass-gen config [set KEY VALUE | reset | path]")
		}
		key, value := args[2], args[3]
		if key == "length" {
			c.Length, err = strconv.Atoi(value)
		} else {
			fields := map[string]*bool{"uppercase": &c.Uppercase, "lowercase": &c.Lowercase, "digits": &c.Digits, "symbols": &c.Symbols, "exclude_ambiguous": &c.ExcludeAmbiguous}
			field, ok := fields[key]
			if !ok {
				return fmt.Errorf("unknown setting %q", key)
			}
			*field, err = strconv.ParseBool(value)
		}
		if err != nil {
			return fmt.Errorf("invalid value for %s: %w", key, err)
		}
		return saveConfig(path, c)
	}
	c, err := loadConfig(path)
	if err != nil {
		return err
	}
	flags := flag.NewFlagSet("pass-gen", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.IntVar(&c.Length, "length", c.Length, "password length")
	flags.BoolVar(&c.Uppercase, "uppercase", c.Uppercase, "include uppercase")
	flags.BoolVar(&c.Lowercase, "lowercase", c.Lowercase, "include lowercase")
	flags.BoolVar(&c.Digits, "digits", c.Digits, "include digits")
	flags.BoolVar(&c.Symbols, "symbols", c.Symbols, "include symbols")
	flags.BoolVar(&c.ExcludeAmbiguous, "exclude-ambiguous", c.ExcludeAmbiguous, "exclude ambiguous characters")
	noCopy := flags.Bool("no-copy", false, "skip clipboard")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, err := io.WriteString(out, help)
			return err
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected argument %q; see pass-gen --help", flags.Arg(0))
	}
	password, err := generate(c)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, password); err != nil {
		return err
	}
	if !*noCopy {
		if err := copyFn(password); err != nil {
			return fmt.Errorf("password printed, but clipboard copy failed: %w", err)
		}
	}
	return nil
}
func main() {
	if err := run(os.Args[1:], os.Stdout, copyPassword); err != nil {
		fmt.Fprintln(os.Stderr, "pass-gen:", err)
		os.Exit(1)
	}
}
