package ai

import (
	"strings"
	"testing"
)

func TestParseTitleBody(t *testing.T) {
	t.Parallel()

	raw := `TITLE: feat: add AI PR generation
BODY:
## Summary
- Adds OpenRouter integration.

## Testing
- [x] go test ./...`

	title, body, err := parseTitleBody(raw)
	if err != nil {
		t.Fatalf("parseTitleBody() error: %v", err)
	}
	if title != "feat: add AI PR generation" {
		t.Fatalf("unexpected title: %q", title)
	}
	if body == "" {
		t.Fatal("expected non-empty body")
	}
}

func TestParseTitleBodyMissingTitle(t *testing.T) {
	t.Parallel()

	_, _, err := parseTitleBody("BODY:\nhello")
	if err == nil {
		t.Fatal("expected error for missing title")
	}
}

func TestParseTitleBodyMissingBody(t *testing.T) {
	t.Parallel()

	_, _, err := parseTitleBody("TITLE: hello")
	if err == nil {
		t.Fatal("expected error for missing body")
	}
}

func TestParseTitleBodyRemovesTODOLines(t *testing.T) {
	t.Parallel()

	raw := `TITLE: feat: improve logs
BODY:
## Summary
- Added richer status output.
- TODO: add integration tests.

## Testing
- [x] go test ./...
- TBD: manual qa`

	_, body, err := parseTitleBody(raw)
	if err != nil {
		t.Fatalf("parseTitleBody() error: %v", err)
	}
	if strings.Contains(strings.ToUpper(body), "TODO") || strings.Contains(strings.ToUpper(body), "TBD") {
		t.Fatalf("body still contains TODO-like tokens: %q", body)
	}
}
