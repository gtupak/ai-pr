package main

import "testing"

func TestParseCreateOptions_Default(t *testing.T) {
	opts, err := parseCreateOptions([]string{})
	if err != nil {
		t.Fatalf("parseCreateOptions returned error: %v", err)
	}
	if opts.HeadOwner != "" {
		t.Fatalf("expected empty head owner, got %q", opts.HeadOwner)
	}
}

func TestParseCreateOptions_WithHeadOwner(t *testing.T) {
	opts, err := parseCreateOptions([]string{"--head-owner", "gtupak"})
	if err != nil {
		t.Fatalf("parseCreateOptions returned error: %v", err)
	}
	if opts.HeadOwner != "gtupak" {
		t.Fatalf("expected head owner %q, got %q", "gtupak", opts.HeadOwner)
	}
}

func TestParseCreateOptions_UnexpectedArg(t *testing.T) {
	_, err := parseCreateOptions([]string{"extra"})
	if err == nil {
		t.Fatalf("expected parseCreateOptions to fail for unexpected argument")
	}
}
