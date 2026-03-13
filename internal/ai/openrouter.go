package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"aipr/internal/git"
)

const (
	defaultModel = "openai/gpt-4o-mini"
	endpointURL  = "https://openrouter.ai/api/v1/chat/completions"
)

type openRouterRequest struct {
	Model       string             `json:"model"`
	Temperature float64            `json:"temperature"`
	Messages    []openRouterPrompt `json:"messages"`
}

type openRouterPrompt struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func GeneratePRTitleBody(apiKey, baseBranch, headBranch string, commits []git.Commit) (string, string, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return "", "", fmt.Errorf("OpenRouter API key is required to generate PR content")
	}

	model := strings.TrimSpace(os.Getenv("AIPR_OPENROUTER_MODEL"))
	if model == "" {
		model = defaultModel
	}

	userPrompt := buildPrompt(baseBranch, headBranch, commits)
	reqBody := openRouterRequest{
		Model:       model,
		Temperature: 0.2,
		Messages: []openRouterPrompt{
			{
				Role: "system",
				Content: "You write concise, high-signal pull request metadata. " +
					"Return exactly two sections in plain text: first line starts with 'TITLE: ', then a 'BODY:' section.",
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", fmt.Errorf("marshal OpenRouter request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpointURL, bytes.NewReader(payload))
	if err != nil {
		return "", "", fmt.Errorf("create OpenRouter request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com")
	req.Header.Set("X-Title", "aipr")

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("call OpenRouter: %w", err)
	}
	defer resp.Body.Close()

	var decoded openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", "", fmt.Errorf("decode OpenRouter response: %w", err)
	}

	if resp.StatusCode >= 300 {
		if decoded.Error != nil && strings.TrimSpace(decoded.Error.Message) != "" {
			return "", "", fmt.Errorf("OpenRouter error (%d): %s", resp.StatusCode, decoded.Error.Message)
		}
		return "", "", fmt.Errorf("OpenRouter error status: %d", resp.StatusCode)
	}

	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return "", "", fmt.Errorf("OpenRouter returned no message content")
	}

	title, body, err := parseTitleBody(decoded.Choices[0].Message.Content)
	if err != nil {
		return "", "", err
	}
	return title, body, nil
}

func buildPrompt(baseBranch, headBranch string, commits []git.Commit) string {
	var b strings.Builder
	b.WriteString("Create a pull request title and body.\n")
	b.WriteString("Constraints:\n")
	b.WriteString("- Keep title under 72 chars.\n")
	b.WriteString("- Body should include sections: Summary and Testing.\n")
	b.WriteString("- Testing section may include TODO checklist if unknown.\n")
	b.WriteString("- No markdown code fences.\n\n")
	b.WriteString("Repository context:\n")
	b.WriteString("- Base branch: " + baseBranch + "\n")
	b.WriteString("- Head branch: " + headBranch + "\n\n")
	b.WriteString("Commits (oldest to newest):\n")
	for i, c := range commits {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, sanitize(c.Subject)))
		if body := sanitize(c.Body); body != "" {
			b.WriteString("   " + body + "\n")
		}
	}
	b.WriteString("\nReturn format:\n")
	b.WriteString("TITLE: <one line title>\n")
	b.WriteString("BODY:\n")
	b.WriteString("<markdown body>\n")
	return b.String()
}

func parseTitleBody(raw string) (string, string, error) {
	text := stripFences(strings.TrimSpace(raw))
	lines := strings.Split(text, "\n")

	title := ""
	bodyStart := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(trimmed), "TITLE:") {
			title = strings.TrimSpace(trimmed[len("TITLE:"):])
		}
		if strings.EqualFold(trimmed, "BODY:") {
			bodyStart = i + 1
			break
		}
	}

	if strings.TrimSpace(title) == "" {
		return "", "", fmt.Errorf("OpenRouter response missing TITLE")
	}
	if bodyStart == -1 || bodyStart > len(lines) {
		return "", "", fmt.Errorf("OpenRouter response missing BODY section")
	}

	body := strings.TrimSpace(strings.Join(lines[bodyStart:], "\n"))
	if body == "" {
		return "", "", fmt.Errorf("OpenRouter response BODY is empty")
	}
	return title, body, nil
}

func stripFences(s string) string {
	trim := strings.TrimSpace(s)
	if strings.HasPrefix(trim, "```") && strings.HasSuffix(trim, "```") {
		trim = strings.TrimPrefix(trim, "```")
		trim = strings.TrimSuffix(trim, "```")
		trim = strings.TrimSpace(trim)
		parts := strings.SplitN(trim, "\n", 2)
		if len(parts) == 2 {
			// Drop optional language tag on the first line.
			if !strings.Contains(parts[0], ":") && len(parts[0]) < 20 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return trim
}

func sanitize(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
}
