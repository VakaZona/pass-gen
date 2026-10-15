package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestTerminalPasswordAndClipboard(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var out bytes.Buffer
	var copied string
	err := runWithUI(nil, &out, func(p string) error { copied = p; return nil }, terminalUI{true, true})
	if err != nil {
		t.Fatal(err)
	}
	if len(copied) != defaults().Length || strings.ContainsAny(copied, "\x1b\n") {
		t.Fatal("clipboard received formatted text")
	}
	for _, want := range []string{copied, "Copied to clipboard", "20 characters", "\x1b["} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q", want)
		}
	}
	out.Reset()
	err = runWithUI(nil, &out, func(string) error { return errors.New("clipboard unavailable") }, terminalUI{true, false})
	if err == nil || strings.Contains(out.String(), "Copied to clipboard") {
		t.Fatal("claimed a failed copy succeeded")
	}
}
func TestPlainOverride(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, args := range [][]string{{"--plain", "--no-copy"}, {"--no-copy", "--plain"}} {
		var out bytes.Buffer
		if err := runWithUI(args, &out, nil, terminalUI{true, true}); err != nil {
			t.Fatal(err)
		}
		if len(out.String()) != 21 || strings.Contains(out.String(), "\x1b") {
			t.Fatal("plain output was decorated")
		}
	}
}
func TestTerminalSettings(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var out bytes.Buffer
	ui := terminalUI{true, false}
	if err := runWithUI([]string{"config", "set", "symbols", "false"}, &out, nil, ui); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "symbols → false") {
		t.Fatal("missing save confirmation")
	}
	out.Reset()
	if err := runWithUI([]string{"config"}, &out, nil, ui); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Symbols                ○ off") || strings.Contains(out.String(), "\x1b") {
		t.Fatal("wrong settings display")
	}
	out.Reset()
	if err := runWithUI([]string{"config", "set", "length", "0"}, &out, nil, ui); err == nil {
		t.Fatal("accepted invalid setting")
	}
	if strings.Contains(out.String(), "Saved") {
		t.Fatal("claimed invalid settings were saved")
	}
	for _, args := range [][]string{{"config", "--json"}, {"--plain", "config"}} {
		out.Reset()
		if err := runWithUI(args, &out, nil, ui); err != nil {
			t.Fatal(err)
		}
		var c Config
		if err := json.Unmarshal(out.Bytes(), &c); err != nil {
			t.Fatal(err)
		}
		if c.Symbols {
			t.Fatal("incorrect JSON settings")
		}
	}
	out.Reset()
	if err := runWithUI([]string{"config", "reset"}, &out, nil, ui); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Defaults restored") {
		t.Fatal("missing reset confirmation")
	}
}
