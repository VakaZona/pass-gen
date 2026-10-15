package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type terminalUI struct{ pretty, color bool }

func newTerminalUI(out io.Writer) terminalUI {
	f, ok := out.(*os.File)
	if !ok || os.Getenv("TERM") == "dumb" {
		return terminalUI{}
	}
	info, err := f.Stat()
	pretty := err == nil && info.Mode()&os.ModeCharDevice != 0
	_, noColor := os.LookupEnv("NO_COLOR")
	return terminalUI{pretty: pretty, color: pretty && !noColor}
}
func (u terminalUI) paint(code, text string) string {
	if !u.color {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}
func (u terminalUI) header(title string) string {
	return "\n  " + u.paint("1", "pass-gen") + "  " + u.paint("2", title) + "\n\n"
}
func (u terminalUI) password(out io.Writer, password string) error {
	if !u.pretty {
		_, err := fmt.Fprintln(out, password)
		return err
	}
	var b strings.Builder
	b.WriteString(u.header("new password"))
	// Short lines keep long passwords readable in narrow terminal windows.
	for start := 0; start < len(password); start += 48 {
		end := start + 48
		if end > len(password) {
			end = len(password)
		}
		fmt.Fprintf(&b, "  %s %s\n", u.paint("2", "│"), u.paint("1;36", password[start:end]))
	}
	b.WriteByte('\n')
	_, err := io.WriteString(out, b.String())
	return err
}
func (u terminalUI) status(out io.Writer, copied bool, length int, c Config) error {
	if !u.pretty {
		return nil
	}
	status := u.paint("2", "Clipboard skipped")
	if copied {
		status = u.paint("32", "✓ Copied to clipboard")
	}
	var categories []string
	if c.Lowercase {
		categories = append(categories, "a–z")
	}
	if c.Uppercase {
		categories = append(categories, "A–Z")
	}
	if c.Digits {
		categories = append(categories, "0–9")
	}
	if c.Symbols {
		categories = append(categories, "symbols")
	}
	_, err := fmt.Fprintf(out, "  %s  %s\n  %s\n\n", status, u.paint("2", fmt.Sprintf("· %d characters", length)), u.paint("2", strings.Join(categories, " · ")))
	return err
}
func (u terminalUI) settings(out io.Writer, c Config, path string) error {
	var b strings.Builder
	b.WriteString(u.header("settings"))
	fmt.Fprintf(&b, "  %-22s %s\n", "Length", u.paint("1;36", strconv.Itoa(c.Length))+" characters")
	for _, row := range []struct {
		label   string
		enabled bool
	}{
		{"Lowercase letters", c.Lowercase}, {"Uppercase letters", c.Uppercase},
		{"Digits", c.Digits}, {"Symbols", c.Symbols}, {"Exclude ambiguous", c.ExcludeAmbiguous},
	} {
		value := u.paint("2", "○ off")
		if row.enabled {
			value = u.paint("32", "● on")
		}
		fmt.Fprintf(&b, "  %-22s %s\n", row.label, value)
	}
	fmt.Fprintf(&b, "\n  %s\n  %s\n\n", u.paint("2", path), u.paint("2", "Change: pass-gen config set <key> <value>"))
	_, err := io.WriteString(out, b.String())
	return err
}
func (u terminalUI) saved(out io.Writer, message, path string) error {
	if !u.pretty {
		return nil
	}
	_, err := fmt.Fprintf(out, "%s  %s %s\n  %s\n\n", u.header("settings"), u.paint("32", "✓"), message, u.paint("2", "Saved to "+path))
	return err
}
func configValue(c Config, key string) string {
	switch key {
	case "length":
		return strconv.Itoa(c.Length)
	case "lowercase":
		return strconv.FormatBool(c.Lowercase)
	case "uppercase":
		return strconv.FormatBool(c.Uppercase)
	case "digits":
		return strconv.FormatBool(c.Digits)
	case "symbols":
		return strconv.FormatBool(c.Symbols)
	case "exclude_ambiguous":
		return strconv.FormatBool(c.ExcludeAmbiguous)
	}
	return ""
}
func (u terminalUI) help(out io.Writer) error {
	if !u.pretty {
		_, err := io.WriteString(out, help)
		return err
	}
	text := strings.TrimPrefix(help, "pass-gen — generate a password and copy it to the clipboard\n\n")
	for _, heading := range []string{"Usage:", "Options (override saved settings for this run):", "Config keys:"} {
		text = strings.ReplaceAll(text, heading, u.paint("1;36", heading))
	}
	_, err := fmt.Fprint(out, u.header("generate · copy · paste"), text, "\n")
	return err
}
