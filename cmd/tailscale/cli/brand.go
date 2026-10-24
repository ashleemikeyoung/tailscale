// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

package cli

import (
	"io"
	"strings"
)

// ElRoi Pneuma branding (Qavats). This client is built from Tailscale's
// open-source code and joins the ElRoi Pneuma tailnet run by Headscale.
// Only what people read is renamed: the command name, help and messages.
// Protocols, file paths, the daemon's service label and API names keep
// their upstream names so that upstream releases merge cleanly.
const (
	// brandProduct is the product name shown to people.
	brandProduct = "ElRoi Pneuma"
	// brandShort is the product name in running text ("Pneuma is starting").
	brandShort = "Pneuma"
	// brandCommand is the name of this CLI.
	brandCommand = "pneuma"
	// brandCredit names the upstream project, as its BSD-3-Clause licence asks.
	brandCredit = "Built on Tailscale (https://github.com/tailscale/tailscale), BSD-3-Clause."
)

// brandText rewrites what people read: command references, so that
// `tailscale up` reads `pneuma up`, and the product name, so that
// "Tailscale is starting" reads "Pneuma is starting" (see [brandProse]). A reference is "tailscale " at the start
// of the text or after whitespace, a quote, a backtick or "(". That leaves
// alone "Tailscale" (the company or protocol), "tailscaled", and hostnames
// such as "pkgs.tailscale.com". (No regexp: tailscaled's dependency test
// keeps it out of the binary.)
func brandText(s string) string {
	return brandProse(brandCommands(s))
}

// brandCommands rewrites command references: "tailscale <subcommand>".
func brandCommands(s string) string {
	const ref = "tailscale "
	if !strings.Contains(s, ref) {
		return s
	}
	var b strings.Builder
	for {
		i := strings.Index(s, ref)
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		b.WriteString(s[:i])
		if i == 0 || strings.ContainsRune(" \t\n'\"`(", rune(s[i-1])) {
			b.WriteString(brandCommand + " ")
		} else {
			b.WriteString(ref)
		}
		s = s[i+len(ref):]
	}
}

// brandErr applies [brandText] to an error's message for display.
func brandErr(err error) error {
	if err == nil {
		return nil
	}
	msg := brandText(err.Error())
	if msg == err.Error() {
		return err
	}
	return &brandedError{msg: msg, err: err}
}

type brandedError struct {
	msg string
	err error
}

func (e *brandedError) Error() string { return e.msg }
func (e *brandedError) Unwrap() error { return e.err }

// brandProse replaces the product name "Tailscale" with [brandShort] in
// running text. It leaves alone the name inside identifiers and paths
// ("TailscaleIPs", "Tailscale.app", "tailscale.com"), the upstream credit
// ("Built on Tailscale") and the company ("Tailscale Inc").
func brandProse(s string) string {
	const name = "Tailscale"
	if !strings.Contains(s, name) {
		return s
	}
	var b strings.Builder
	for {
		i := strings.Index(s, name)
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		before, after := b.String()+s[:i], s[i+len(name):]
		b.WriteString(s[:i])
		if isProductMention(before, after) {
			b.WriteString(brandShort)
		} else {
			b.WriteString(name)
		}
		s = after
	}
}

func isProductMention(before, after string) bool {
	if before != "" {
		if c := before[len(before)-1]; isWordByte(c) || strings.IndexByte("./-_@", c) >= 0 {
			return false
		}
	}
	if after != "" {
		c := after[0]
		if isWordByte(c) {
			return false
		}
		if c == '.' && len(after) > 1 && isWordByte(after[1]) {
			return false // Tailscale.app, Tailscale.exe
		}
	}
	if strings.HasSuffix(before, "Built on ") || strings.HasPrefix(after, " Inc") || strings.HasPrefix(after, ", Inc") {
		return false
	}
	return true
}

func isWordByte(c byte) bool {
	return c == '_' || '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

// brandWriter applies [brandText] to what the CLI prints for people. JSON
// documents (a write that starts with '{' or '[') pass through unchanged, so
// scripts reading --json output see upstream's values.
type brandWriter struct{ w io.Writer }

func (bw brandWriter) Write(p []byte) (int, error) {
	if t := strings.TrimLeft(string(p), " \t\r\n"); strings.HasPrefix(t, "{") || strings.HasPrefix(t, "[") {
		return bw.w.Write(p)
	}
	if _, err := io.WriteString(bw.w, brandText(string(p))); err != nil {
		return 0, err
	}
	return len(p), nil
}

// branded wraps w in a [brandWriter] once.
func branded(w io.Writer) io.Writer {
	if _, ok := w.(brandWriter); ok {
		return w
	}
	return brandWriter{w}
}

// unbranded returns the writer under a [brandWriter].
func unbranded(w io.Writer) io.Writer {
	if bw, ok := w.(brandWriter); ok {
		return bw.w
	}
	return w
}
