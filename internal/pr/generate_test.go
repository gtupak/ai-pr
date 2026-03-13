package pr

import (
	"strings"
	"testing"

	"aipr/internal/git"
)

func TestGenerateTitleNoCommits(t *testing.T) {
	t.Parallel()

	got := GenerateTitle(nil)
	if got != "Update changes" {
		t.Fatalf("unexpected title: %q", got)
	}
}

func TestGenerateTitleSingleCommit(t *testing.T) {
	t.Parallel()

	commits := []git.Commit{
		{Subject: "feat: add aipr config command"},
	}

	got := GenerateTitle(commits)
	want := "feat: add aipr config command"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestGenerateTitleMultipleCommits(t *testing.T) {
	t.Parallel()

	commits := []git.Commit{
		{Subject: "feat: add cli bootstrap"},
		{Subject: "test: add config tests"},
		{Subject: "docs: add README"},
	}

	got := GenerateTitle(commits)
	want := "feat: add cli bootstrap (+2 more commits)"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestGenerateBodyIncludesCommitSubjects(t *testing.T) {
	t.Parallel()

	commits := []git.Commit{
		{Subject: "feat: add base branch config"},
		{Subject: "fix: handle detached head"},
	}

	got := GenerateBody(commits)
	if !strings.Contains(got, "## Summary") {
		t.Fatalf("body missing summary section: %q", got)
	}
	if !strings.Contains(got, "- feat: add base branch config") {
		t.Fatalf("body missing first commit: %q", got)
	}
	if !strings.Contains(got, "- fix: handle detached head") {
		t.Fatalf("body missing second commit: %q", got)
	}
}
