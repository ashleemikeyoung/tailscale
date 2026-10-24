// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

package cli

import "strings"

// ElRoi Pneuma branding (Qavats). This client is built from Tailscale's
// open-source code and joins the ElRoi Pneuma tailnet run by Headscale.
// Only what people read is renamed: the command name, help and messages.
// Protocols, file paths, the daemon's service label and API names keep
// their upstream names so that upstream releases merge cleanly.
const (
	// brandProduct is the product name shown to people.
	brandProduct = "ElRoi Pneuma"
	// brandCommand is the name of this CLI.
	brandCommand = "pneuma"
	// brandCredit names the upstream project, as its BSD-3-Clause licence asks.
	brandCredit = "Built on Tailscale (https://github.com/tailscale/tailscale), BSD-3-Clause."
)

// brandText rewrites command references in help and messages, so that
// `tailscale up` reads `pneuma up`. A reference is "tailscale " at the start
// of the text or after whitespace, a quote, a backtick or "(". That leaves
// alone "Tailscale" (the company or protocol), "tailscaled", and hostnames
// such as "pkgs.tailscale.com". (No regexp: tailscaled's dependency test
// keeps it out of the binary.)
func brandText(s string) string {
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
