// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

package cli

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestBrandText(t *testing.T) {
	tests := []struct{ in, want string }{
		{"tailscale up [flags]", "pneuma up [flags]"},
		{`see "tailscale set --help" for valid settings`, `see "pneuma set --help" for valid settings`},
		{"use `tailscale set --exit-node=`", "use `pneuma set --exit-node=`"},
		{"Use 'sudo tailscale up'.", "Use 'sudo pneuma up'."},
		{"(tailscale status)", "(pneuma status)"},
		// the product name in running text
		{"Tailscale is starting. Please wait.", "Pneuma is starting. Please wait."},
		{"Tailscale is stopped.", "Pneuma is stopped."},
		{"Tailscale SSH", "Pneuma SSH"},
		{"Tailscale DNS: enabled.", "Pneuma DNS: enabled."},
		{"could not reach the Tailscale coordination server", "could not reach the Pneuma coordination server"},
		{"Connect to Tailscale, logging in if needed", "Connect to Pneuma, logging in if needed"},
		{"Run 'tailscale up' to start Tailscale.", "Run 'pneuma up' to start Pneuma."},
		// left alone: identifiers, paths, URLs, the daemon, the credit and the company
		{"tailscaled --state=x", "tailscaled --state=x"},
		{"https://pkgs.tailscale.com/stable", "https://pkgs.tailscale.com/stable"},
		{"tailscale.com/cmd/tailscale", "tailscale.com/cmd/tailscale"},
		{"TailscaleIPs", "TailscaleIPs"},
		{"/Applications/Tailscale.app", "/Applications/Tailscale.app"},
		{brandCredit, brandCredit},
		{"Copyright (c) Tailscale Inc & contributors", "Copyright (c) Tailscale Inc & contributors"},
	}
	for _, tt := range tests {
		if got := brandText(tt.in); got != tt.want {
			t.Errorf("brandText(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestBrandErrKeepsTheChain(t *testing.T) {
	base := errors.New("final argument to 'tailscale file cp' must end in colon")
	err := brandErr(base)
	if !strings.Contains(err.Error(), "'pneuma file cp'") {
		t.Errorf("message not branded: %q", err)
	}
	if !errors.Is(err, base) {
		t.Error("errors.Is no longer finds the original error")
	}
	if plain := errors.New("no command here"); brandErr(plain) != plain {
		t.Error("an error without command references should be returned as is")
	}
	if brandErr(nil) != nil {
		t.Error("nil must stay nil")
	}
}

func TestRootHelpIsBranded(t *testing.T) {
	root := newRootCmd()
	help := usageFunc(root)
	for _, want := range []string{"pneuma [flags] <subcommand>", brandProduct, brandCredit} {
		if !strings.Contains(help, want) {
			t.Errorf("root help missing %q:\n%s", want, help)
		}
	}
	for _, sub := range root.Subcommands {
		if sub.Name == "up" {
			if h := usageFunc(sub); !strings.Contains(h, "pneuma up [flags]") || strings.Contains(h, "tailscale up") {
				t.Errorf("up help not branded:\n%s", h)
			}
		}
	}
}

func TestBrandWriter(t *testing.T) {
	var buf strings.Builder
	w := branded(&buf)
	if branded(w) != w {
		t.Error("branded must not wrap twice")
	}
	if unbranded(w) != &buf {
		t.Error("unbranded must return the underlying writer")
	}
	fmt.Fprintln(w, "Tailscale is starting. Please wait.")
	fmt.Fprintln(w, `{"Health": ["Tailscale is starting. Please wait."]}`)
	got := buf.String()
	want := "Pneuma is starting. Please wait.\n" + `{"Health": ["Tailscale is starting. Please wait."]}` + "\n"
	if got != want {
		t.Errorf("got %q, want %q (JSON passes through unchanged)", got, want)
	}
}
