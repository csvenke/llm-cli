package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestUsageIncludesGHPR(t *testing.T) {
	var out bytes.Buffer
	usageTo(&out)

	if !strings.Contains(out.String(), "gh <subcommand>") {
		t.Fatalf("usage output = %q, want to include %q", out.String(), "gh <subcommand>")
	}

	if !strings.Contains(out.String(), "pr                Create a GitHub pull request") {
		t.Fatalf("usage output = %q, want to include gh pr subcommand help", out.String())
	}
}
