package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	for _, c := range []Config{defaults(), {4, true, true, true, true, true}, {32, false, false, true, false, false}} {
		groups, _ := c.groups()
		for i := 0; i < 100; i++ {
			p, err := generate(c)
			if err != nil {
				t.Fatal(err)
			}
			if len(p) != c.Length {
				t.Fatalf("wrong length %d", len(p))
			}
			for _, group := range groups {
				if !strings.ContainsAny(p, group) {
					t.Fatal("missing category")
				}
			}
			for _, ch := range p {
				if !strings.ContainsRune(strings.Join(groups, ""), ch) {
					t.Fatal("unexpected character")
				}
			}
		}
	}
}
func TestInvalidConfig(t *testing.T) {
	for _, c := range []Config{{}, {0, true, true, true, true, true}, {3, true, true, true, true, true}, {4097, true, true, true, true, true}} {
		if _, err := generate(c); err == nil {
			t.Fatal("accepted invalid config")
		}
	}
}
func TestSettingsAndClipboard(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var out bytes.Buffer
	var copied string
	copyFn := func(s string) error { copied = s; return nil }
	if err := run([]string{"config", "set", "length", "30"}, &out, copyFn); err != nil {
		t.Fatal(err)
	}
	if err := run(nil, &out, copyFn); err != nil {
		t.Fatal(err)
	}
	if len(copied) != 30 || out.String() != copied+"\n" {
		t.Fatal("print and clipboard differ")
	}
	out.Reset()
	if err := run([]string{"--length", "12", "--no-copy"}, &out, func(string) error { t.Fatal("copied despite --no-copy"); return nil }); err != nil {
		t.Fatal(err)
	}
	if len(strings.TrimSpace(out.String())) != 12 {
		t.Fatal("override failed")
	}
	path, _ := configPath()
	c, _ := loadConfig(path)
	if c.Length != 30 {
		t.Fatal("override persisted")
	}
	if err := run([]string{"config", "set", "length", "0"}, &out, copyFn); err == nil {
		t.Fatal("saved invalid config")
	}
	c, _ = loadConfig(path)
	if c.Length != 30 {
		t.Fatal("invalid update changed config")
	}
	if err := run([]string{"config", "reset"}, &out, copyFn); err != nil {
		t.Fatal(err)
	}
	c, _ = loadConfig(path)
	if c != defaults() {
		t.Fatal("reset failed")
	}
}
func TestCopyFailure(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var out bytes.Buffer
	err := run(nil, &out, func(string) error { return errors.New("unavailable") })
	if err == nil || !strings.Contains(err.Error(), "clipboard") || out.Len() == 0 {
		t.Fatal("clipboard failure not reported")
	}
}
func TestMalformedConfigAndRecovery(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path, _ := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run(nil, &bytes.Buffer{}, nil); err == nil {
		t.Fatal("accepted malformed config")
	}
	if err := run([]string{"config", "reset"}, &bytes.Buffer{}, nil); err != nil {
		t.Fatal(err)
	}
}
