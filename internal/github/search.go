package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

type SearchResult struct {
	APIKey   string
	RepoName string
	FilePath string
}

type GithubSearcher struct {
	client *github.Client
}

func NewGithubSearcher(token string) *GithubSearcher {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &GithubSearcher{
		client: client,
	}
}

func (gs *GithubSearcher) Search(keyword string) ([]SearchResult, error) {
	ctx := context.Background()
	// Try different search patterns
	queries := []string{
		fmt.Sprintf(`%s path:**/.env`, keyword),
		fmt.Sprintf(`%s filename:.env`, keyword),
		fmt.Sprintf(`%s in:file filename:.env`, keyword),
	}

	var allResults []SearchResult

	// Test the GitHub token first
	_, _, err := gs.client.Users.Get(ctx, "")
	if err != nil {
		log.Printf("GitHub token validation error: %v", err)
		return nil, fmt.Errorf("invalid GitHub token or insufficient permissions: %v", err)
	}

	for _, query := range queries {
		log.Printf("Trying search query: %s", query)

		opts := &github.SearchOptions{
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		}

		for {
			result, resp, err := gs.client.Search.Code(ctx, query, opts)
			if err != nil {
				if _, ok := err.(*github.RateLimitError); ok {
					log.Println("Hit GitHub rate limit")
					return nil, fmt.Errorf("GitHub rate limit exceeded")
				}
				log.Printf("Search error for query '%s': %v", query, err)
				// Continue with next query instead of returning
				break
			}

			if result.GetTotal() == 0 {
				log.Printf("No results found for query: %s", query)
				break
			}

			for _, item := range result.CodeResults {
				log.Printf("Found potential match in: %s/%s", *item.Repository.FullName, *item.Path)
				content, err := gs.getFileContent(ctx, *item.Repository.Owner.Login, *item.Repository.Name, *item.Path)
				if err != nil {
					log.Printf("Error getting content: %v", err)
					continue
				}

				apiKeys := extractAPIKeys(content, keyword)
				for _, apiKey := range apiKeys {
					log.Printf("Found potential API key in %s", *item.Repository.FullName)
					allResults = append(allResults, SearchResult{
						APIKey:   apiKey,
						RepoName: *item.Repository.FullName,
						FilePath: *item.Path,
					})
				}
			}

			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
	}

	return allResults, nil
}

func (gs *GithubSearcher) getFileContent(ctx context.Context, owner, repo, path string) (string, error) {
	content, _, _, err := gs.client.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		return "", err
	}

	decoded, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

func extractAPIKeys(content, keyword string) []string {
	var keys []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		// Skip comments and empty lines
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		lowerLine := strings.ToLower(line)
		lowerKeyword := strings.ToLower(keyword)
		// Check for both the exact keyword and common variations
		if strings.Contains(lowerLine, lowerKeyword) ||
			strings.Contains(lowerLine, strings.Replace(lowerKeyword, "_", "", -1)) ||
			strings.Contains(lowerLine, strings.Replace(lowerKeyword, "_api", "", -1)) {

			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[1])
				// Remove quotes and any trailing comments
				key = strings.Split(key, "#")[0]
				key = strings.Trim(key, "\"'`")
				key = strings.TrimSpace(key)
				// Less strict validation
				if len(key) > 10 && isValidAPIKey(key) {
					keys = append(keys, key)
				}
			}
		}
	}

	return keys
}

func isValidAPIKey(key string) bool {
	if len(key) < 10 || len(key) > 200 {
		return false
	}

	hasLetter := false
	hasNumber := false
	hasSpecial := false
	for _, char := range key {
		if unicode.IsLetter(char) {
			hasLetter = true
		}
		if unicode.IsNumber(char) {
			hasNumber = true
		}
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
			hasSpecial = true
		}
	}

	// Allow keys that have at least two of: letters, numbers, special chars
	count := 0
	if hasLetter {
		count++
	}
	if hasNumber {
		count++
	}
	if hasSpecial {
		count++
	}
	return count >= 2
}
