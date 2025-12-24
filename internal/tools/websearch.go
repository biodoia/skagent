package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// WebSearchTool provides web search capabilities
type WebSearchTool struct {
	httpClient *http.Client
	timeout    time.Duration
}

// NewWebSearchTool creates a new web search tool
func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		timeout:    15 * time.Second,
	}
}

// Name returns the tool identifier
func (w *WebSearchTool) Name() string {
	return "websearch"
}

// Description returns tool description
func (w *WebSearchTool) Description() string {
	return "Search the web for information, GitHub repositories, and documentation"
}

// CanHandle checks if this tool can handle the intent
func (w *WebSearchTool) CanHandle(intent string) bool {
	lower := strings.ToLower(intent)
	keywords := []string{"search", "find", "look up", "lookup", "google", "web"}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// Execute performs a web search
func (w *WebSearchTool) Execute(ctx context.Context, input string) (string, error) {
	lower := strings.ToLower(input)

	// Add timeout to context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, w.timeout)
		defer cancel()
	}

	// Determine search type
	if strings.Contains(lower, "github") || strings.Contains(lower, "repo") {
		return w.searchGitHub(ctx, input)
	}

	// Default to DuckDuckGo Instant Answer
	return w.searchDuckDuckGo(ctx, input)
}

// searchGitHub searches GitHub repositories
func (w *WebSearchTool) searchGitHub(ctx context.Context, query string) (string, error) {
	// Extract search terms (remove common words)
	terms := extractSearchTerms(query)
	if len(terms) == 0 {
		return "", fmt.Errorf("no search terms found")
	}

	searchQuery := url.QueryEscape(strings.Join(terms, " "))
	apiURL := fmt.Sprintf("https://api.github.com/search/repositories?q=%s&sort=stars&per_page=5", searchQuery)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "SkAgent/1.0")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("search timed out")
		}
		return "", fmt.Errorf("search failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		TotalCount int `json:"total_count"`
		Items      []struct {
			FullName    string `json:"full_name"`
			Description string `json:"description"`
			HTMLURL     string `json:"html_url"`
			Stars       int    `json:"stargazers_count"`
			Language    string `json:"language"`
		} `json:"items"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Items) == 0 {
		return fmt.Sprintf("No GitHub repositories found for: %s", strings.Join(terms, " ")), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d repositories for '%s':\n\n", result.TotalCount, strings.Join(terms, " ")))

	for i, repo := range result.Items {
		sb.WriteString(fmt.Sprintf("%d. **%s** ⭐ %d\n", i+1, repo.FullName, repo.Stars))
		if repo.Description != "" {
			desc := repo.Description
			if len(desc) > 100 {
				desc = desc[:100] + "..."
			}
			sb.WriteString(fmt.Sprintf("   %s\n", desc))
		}
		if repo.Language != "" {
			sb.WriteString(fmt.Sprintf("   Language: %s\n", repo.Language))
		}
		sb.WriteString(fmt.Sprintf("   %s\n\n", repo.HTMLURL))
	}

	return sb.String(), nil
}

// searchDuckDuckGo uses DuckDuckGo Instant Answer API
func (w *WebSearchTool) searchDuckDuckGo(ctx context.Context, query string) (string, error) {
	terms := extractSearchTerms(query)
	if len(terms) == 0 {
		return "", fmt.Errorf("no search terms found")
	}

	searchQuery := url.QueryEscape(strings.Join(terms, " "))
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1", searchQuery)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "SkAgent/1.0")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("search timed out")
		}
		return "", fmt.Errorf("search failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Abstract     string `json:"Abstract"`
		AbstractText string `json:"AbstractText"`
		AbstractURL  string `json:"AbstractURL"`
		Heading      string `json:"Heading"`
		RelatedTopics []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"RelatedTopics"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Search results for '%s':\n\n", strings.Join(terms, " ")))

	if result.AbstractText != "" {
		sb.WriteString(fmt.Sprintf("**%s**\n", result.Heading))
		sb.WriteString(fmt.Sprintf("%s\n", result.AbstractText))
		if result.AbstractURL != "" {
			sb.WriteString(fmt.Sprintf("Source: %s\n\n", result.AbstractURL))
		}
	}

	if len(result.RelatedTopics) > 0 {
		sb.WriteString("Related:\n")
		count := 0
		for _, topic := range result.RelatedTopics {
			if topic.Text != "" && count < 5 {
				text := topic.Text
				if len(text) > 150 {
					text = text[:150] + "..."
				}
				sb.WriteString(fmt.Sprintf("• %s\n", text))
				count++
			}
		}
	}

	if sb.Len() < 50 {
		return fmt.Sprintf("No detailed results found for '%s'. Try searching on GitHub or using more specific terms.", strings.Join(terms, " ")), nil
	}

	return sb.String(), nil
}

// extractSearchTerms removes common words from search query
func extractSearchTerms(query string) []string {
	stopWords := map[string]bool{
		"search": true, "find": true, "look": true, "up": true, "for": true,
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"on": true, "in": true, "to": true, "of": true, "and": true,
		"github": true, "repo": true, "repository": true, "web": true,
		"please": true, "can": true, "you": true, "me": true,
	}

	words := strings.Fields(strings.ToLower(query))
	var terms []string
	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?\"'")
		if word != "" && !stopWords[word] && len(word) > 1 {
			terms = append(terms, word)
		}
	}
	return terms
}
