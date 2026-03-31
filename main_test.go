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
	if opts.Base != "" {
		t.Fatalf("expected empty base, got %q", opts.Base)
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
	if opts.Base != "" {
		t.Fatalf("expected empty base, got %q", opts.Base)
	}
}

func TestParseCreateOptions_WithBase(t *testing.T) {
	opts, err := parseCreateOptions([]string{"--base", "develop"})
	if err != nil {
		t.Fatalf("parseCreateOptions returned error: %v", err)
	}
	if opts.Base != "develop" {
		t.Fatalf("expected base %q, got %q", "develop", opts.Base)
	}
}

func TestParseCreateOptions_WithBaseAndHeadOwner(t *testing.T) {
	opts, err := parseCreateOptions([]string{"--base", "main", "--head-owner", "forkuser"})
	if err != nil {
		t.Fatalf("parseCreateOptions returned error: %v", err)
	}
	if opts.Base != "main" {
		t.Fatalf("expected base %q, got %q", "main", opts.Base)
	}
	if opts.HeadOwner != "forkuser" {
		t.Fatalf("expected head owner %q, got %q", "forkuser", opts.HeadOwner)
	}
}

func TestParseCreateOptions_UnexpectedArg(t *testing.T) {
	_, err := parseCreateOptions([]string{"extra"})
	if err == nil {
		t.Fatalf("expected parseCreateOptions to fail for unexpected argument")
	}
}
