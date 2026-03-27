package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	githubAPIURL = "https://api.github.com"
	githubURL    = "https://github.com"
)

// Client handles GitHub API interactions
type Client struct {
	tokens     []string
	tokenIndex int
	httpClient *http.Client
	mu         sync.Mutex
}

// TokenUsage tracks token usage per user
type TokenUsage struct {
	Count     int
	ResetTime time.Time
}

// NewClient creates a new GitHub client
func NewClient(tokens []string) *Client {
	return &Client{
		tokens:     tokens,
		tokenIndex: 0,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetTokens sets the GitHub tokens
func (c *Client) SetTokens(tokens []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokens = tokens
	c.tokenIndex = 0
}

// getToken returns the current token and rotates to next
func (c *Client) getToken() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.tokens) == 0 {
		return ""
	}

	token := c.tokens[c.tokenIndex]
	// Rotate to next token
	c.tokenIndex = (c.tokenIndex + 1) % len(c.tokens)
	return token
}

// doRequest makes an authenticated request to GitHub API
func (c *Client) doRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	token := c.getToken()
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return c.httpClient.Do(req)
}

// UserStats holds GitHub user statistics
type UserStats struct {
	Name              string `json:"name"`
	Login             string `json:"login"`
	PublicRepos       int    `json:"public_repos"`
	Followers         int    `json:"followers"`
	Following         int    `json:"following"`
	CreatedAt         string `json:"created_at"`
	Bio               string `json:"bio"`
	AvatarURL         string `json:"avatar_url"`
	TotalStars        int
	TotalCommits      int
	TotalPRs          int
	TotalIssues       int
	TotalContributions int
	TotalRepos        int
	TotalReviews      int
	DiscussionsStarted int
	DiscussionsAnswered int
	PRsMerged         int
	PRsMergedPercentage float64
}

// GetUser fetches user information
func (c *Client) GetUser(username string) (*UserStats, error) {
	url := fmt.Sprintf("%s/users/%s", githubAPIURL, username)
	
	resp, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	
	var user UserStats
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	
	return &user, nil
}

// Repository represents a GitHub repository
type Repository struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Stars       int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	OpenIssues  int    `json:"open_issues_count"`
	IsFork      bool   `json:"fork"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	PushedAt    string `json:"pushed_at"`
	HTMLURL     string `json:"html_url"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
}

// GetRepos fetches user's repositories
func (c *Client) GetRepos(username string, page int) ([]Repository, error) {
	url := fmt.Sprintf("%s/users/%s/repos?per_page=100&page=%d", githubAPIURL, username, page)
	
	resp, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	
	var repos []Repository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}
	
	return repos, nil
}

// GetAllRepos fetches all user repositories
func (c *Client) GetAllRepos(username string) ([]Repository, error) {
	var allRepos []Repository
	page := 1
	
	for {
		repos, err := c.GetRepos(username, page)
		if err != nil {
			return nil, err
		}
		
		if len(repos) == 0 {
			break
		}
		
		allRepos = append(allRepos, repos...)
		
		if len(repos) < 100 {
			break
		}
		
		page++
	}
	
	return allRepos, nil
}

// GetRepo fetches a specific repository
func (c *Client) GetRepo(owner, repo string) (*Repository, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", githubAPIURL, owner, repo)
	
	resp, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	
	var repository Repository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, err
	}
	
	return &repository, nil
}

// LanguageStats holds language statistics
type LanguageStats struct {
	Name      string
	Bytes     int
	Color     string
	Repos     int
	Percentage float64
}

// GetLanguages fetches language statistics for a user
func (c *Client) GetLanguages(username string, excludeRepos []string) (map[string]*LanguageStats, error) {
	repos, err := c.GetAllRepos(username)
	if err != nil {
		return nil, err
	}
	
	languages := make(map[string]*LanguageStats)
	
	for _, repo := range repos {
		// Skip forks
		if repo.IsFork {
			continue
		}
		
		// Skip excluded repos
		skip := false
		for _, exclude := range excludeRepos {
			if repo.Name == exclude {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		
		// Fetch languages for this repo
		url := fmt.Sprintf("%s/repos/%s/%s/languages", githubAPIURL, username, repo.Name)
		
		resp, err := c.doRequest(url)
		if err != nil {
			continue
		}
		
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			continue
		}
		
		var langs map[string]int
		if err := json.Unmarshal(body, &langs); err != nil {
			continue
		}
		
		for lang, bytes := range langs {
			if _, ok := languages[lang]; !ok {
				languages[lang] = &LanguageStats{
					Name:  lang,
					Color: GetLanguageColor(lang),
				}
			}
			languages[lang].Bytes += bytes
			languages[lang].Repos++
		}
	}
	
	return languages, nil
}

// GraphQLResponse for contributions
const contributionsQuery = `
query($login: String!) {
	user(login: $login) {
		contributionsCollection {
			totalCommitContributions
			totalPullRequestContributions
			totalIssueContributions
			totalRepositoryContributions
			pullRequestReviewContributions {
				totalCount
			}
		}
		discussionsStarted: repositoryDiscussions {
			totalCount
		}
		discussionsAnswered: repositoryDiscussionComments(onlyAnswers: true) {
			totalCount
		}
		pullRequests(states: MERGED) {
			totalCount
		}
	}
}
`

// GetContributions fetches user contributions using GraphQL
func (c *Client) GetContributions(username string) (map[string]int, error) {
	// For now, return mock data since GraphQL requires different handling
	// In production, this should use GitHub GraphQL API
	return map[string]int{
		"commits": 0,
		"prs": 0,
		"issues": 0,
		"reviews": 0,
		"discussions_started": 0,
		"discussions_answered": 0,
		"prs_merged": 0,
	}, nil
}

// Gist represents a GitHub Gist
type Gist struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
	Files map[string]struct {
		Filename string `json:"filename"`
		Language string `json:"language"`
		Size     int    `json:"size"`
	} `json:"files"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// GetGist fetches a specific gist
func (c *Client) GetGist(id string) (*Gist, error) {
	url := fmt.Sprintf("%s/gists/%s", githubAPIURL, id)
	
	resp, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	
	var gist Gist
	if err := json.NewDecoder(resp.Body).Decode(&gist); err != nil {
		return nil, err
	}
	
	return &gist, nil
}

// GetStarredRepos gets starred repositories count
func (c *Client) GetStarredRepos(username string) (int, error) {
	url := fmt.Sprintf("%s/users/%s/starred?per_page=1", githubAPIURL, username)
	
	resp, err := c.doRequest(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	// Get count from Link header or count manually
	linkHeader := resp.Header.Get("Link")
	if linkHeader != "" {
		// Parse last page from Link header
		// Link: <...>; rel="last"
		parts := strings.Split(linkHeader, ",")
		for _, part := range parts {
			if strings.Contains(part, `rel="last"`) {
				// Extract page number
				start := strings.Index(part, "page=")
				if start != -1 {
					end := strings.Index(part[start:], ">")
					if end != -1 {
						pageStr := part[start+5 : start+end]
						var count int
						fmt.Sscanf(pageStr, "%d", &count)
						return count, nil
					}
				}
			}
		}
	}
	
	// Fallback: count manually
	var starred []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&starred); err != nil {
		return 0, err
	}
	
	return len(starred), nil
}

// GetLanguageColor returns the color for a programming language
func GetLanguageColor(language string) string {
	colors := map[string]string{
		"JavaScript": "#f1e05a",
		"TypeScript": "#2b7489",
		"Python":     "#3572A5",
		"Java":       "#b07219",
		"Go":         "#00ADD8",
		"Rust":       "#dea584",
		"C++":        "#f34b7d",
		"C":          "#555555",
		"C#":         "#178600",
		"PHP":        "#4F5D95",
		"Ruby":       "#701516",
		"Swift":      "#ffac45",
		"Kotlin":     "#A97BFF",
		"HTML":       "#e34c26",
		"CSS":        "#563d7c",
		"Shell":      "#89e051",
		"Vue":        "#41b883",
		"React":      "#61dafb",
		"Dockerfile": "#384d54",
		"Markdown":   "#083fa1",
	}
	
	if color, ok := colors[language]; ok {
		return color
	}
	return "#858585"
}
