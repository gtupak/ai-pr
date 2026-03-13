package pr

import (
	"fmt"
	"strings"

	"aipr/internal/git"
)

func GenerateTitle(commits []git.Commit) string {
	if len(commits) == 0 {
		return "Update changes"
	}

	first := cleanLine(commits[0].Subject)
	if len(commits) == 1 {
		return fallback(first, "Update changes")
	}

	commitCount := len(commits)
	if first == "" {
		return fmt.Sprintf("Update branch (%d commits)", commitCount)
	}
	return fmt.Sprintf("%s (+%d more commits)", first, commitCount-1)
}

func GenerateBody(commits []git.Commit) string {
	if len(commits) == 0 {
		return "## Summary\n- No commit details available.\n"
	}

	var b strings.Builder
	b.WriteString("## Summary\n")
	b.WriteString(fmt.Sprintf("- Includes %d commit(s) from this branch.\n", len(commits)))
	b.WriteString("- Generated from commit history.\n\n")
	b.WriteString("## Commits\n")
	for _, c := range commits {
		subject := cleanLine(c.Subject)
		if subject == "" {
			continue
		}
		b.WriteString("- ")
		b.WriteString(subject)
		b.WriteString("\n")
	}

	return b.String()
}

func cleanLine(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
}

func fallback(primary, backup string) string {
	if strings.TrimSpace(primary) == "" {
		return backup
	}
	return primary
}
