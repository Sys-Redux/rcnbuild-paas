package github

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	githubAPIBaseURL = "https://api.github.com"
	userAgent        = "RCNbuild-PaaS/1.0"
)

// Client wraps GitHub API calls with auth
type Client struct {
	accessToken string
	httpClient  *http.Client
}

// Creates a GitHub API client with the provided access token
func NewClient(accessToken string) *Client {
	return &Client{
		accessToken: accessToken,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Represents a GitHub repository
type Repository struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Description   string    `json:"description"`
	Private       bool      `json:"private"`
	HTMLURL       string    `json:"html_url"`
	CloneURL      string    `json:"clone_url"`
	SSHURL        string    `json:"ssh_url"`
	DefaultBranch string    `json:"default_branch"`
	Language      string    `json:"language"`
	UpdatedAt     time.Time `json:"updated_at"`
	Permissions   struct {
		Admin bool `json:"admin"`
		Push  bool `json:"push"`
		Pull  bool `json:"pull"`
	} `json:"permissions"`
}

// Represents a GitHub webhook
type Webhook struct {
	ID     int64    `json:"id"`
	Name   string   `json:"name"`
	Active bool     `json:"active"`
	Events []string `json:"events"`
	Config struct {
		URL         string `json:"url"`
		ContentType string `json:"content_type"`
		InsecureSSL string `json:"insecure_ssl"`
	} `json:"config"`
}

// Create a webhook
type WebhookCreateRequest struct {
	Name   string              `json:"name"`
	Active bool                `json:"active"`
	Events []string            `json:"events"`
	Config WebhookCreateConfig `json:"config"`
}

// Contains webhook config
type WebhookCreateConfig struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Secret      string `json:"secret"`
	InsecureSSL string `json:"insecure_ssl"`
}

// Represents a file or directory in a repository
type RepoContent struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"` // "file" or "dir"
}

// Perform an authenticated request to the GitHub API
func (c *Client) doRequest(ctx context.Context, method, endpoint string,
	body io.Reader) (*http.Response, error) {
	url := githubAPIBaseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

// Fetch repositorues for the authenticated user
// Only returns repos where the user has push access
func (c *Client) ListUserRepos(ctx context.Context, page,
	perPage int) ([]*Repository, error) {
	if perPage <= 0 {
		perPage = 30
	}
	if page <= 0 {
		page = 1
	}

	endpoint := fmt.Sprintf("/user/repos?sort=updated&per_page=%d&page=%d&affiliation=owner,collaborator,organization_member", perPage, page)

	resp, err := c.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch repos: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s - %s",
			resp.Status, string(body))
	}

	var repos []*Repository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("Failed to decode repos response: %w", err)
	}

	// Filter only repos where user can push
	var deployableRepos []*Repository
	for _, repo := range repos {
		if repo.Permissions.Push || repo.Permissions.Admin {
			deployableRepos = append(deployableRepos, repo)
		}
	}
	return deployableRepos, nil
}

// Fetch a specific repo by owner/repo
func (c *Client) GetRepo(ctx context.Context, owner,
	repo string) (*Repository, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s", owner, repo)

	resp, err := c.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch repo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("Repository not found: %s/%s", owner, repo)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s - %s",
			resp.Status, string(body))
	}

	var repository Repository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, fmt.Errorf("Failed to decode repo response: %w", err)
	}
	return &repository, nil
}

// Get contents of a directory in a repository
// Used for runtime detection
func (c *Client) GetRepoContents(ctx context.Context, owner,
	repo, path, ref string) ([]*RepoContent, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path)
	if ref != "" {
		endpoint += "?ref=" + ref
	}

	resp, err := c.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch repo contents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("Repository or path not found: %s/%s/%s",
			owner, repo, path)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s - %s",
			resp.Status, string(body))
	}

	var contents []*RepoContent
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, fmt.Errorf(
			"Failed to decode repo contents response: %w", err,
		)
	}
	return contents, nil
}

// Checks if specific file exists in the repository
func (c *Client) FileExists(ctx context.Context, owner, repo,
	path, ref string) (bool, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path)
	if ref != "" {
		endpoint += "?ref=" + ref
	}

	resp, err := c.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false, fmt.Errorf("Failed to check file existence: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// Create a cryptographically secure webhook secret
func GenerateWebhookSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Create a webhook for push events on a repository
func (c *Client) CreateWebhook(ctx context.Context, owner, repo,
	webhookURL, secret string) (*Webhook, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/hooks", owner, repo)

	payload := WebhookCreateRequest{
		Name:   "web",
		Active: true,
		Events: []string{"push", "pull_request"},
		Config: WebhookCreateConfig{
			URL:         webhookURL,
			ContentType: "json",
			Secret:      secret,
			InsecureSSL: "0",
		},
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal webhook payload: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, endpoint,
		strings.NewReader(string(payloadJSON)))
	if err != nil {
		return nil, fmt.Errorf("Failed to create webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Failed to create webhook: %s - %s",
			resp.Status, string(body))
	}

	var webhook Webhook
	if err := json.NewDecoder(resp.Body).Decode(&webhook); err != nil {
		return nil, fmt.Errorf("Failed to decode webhook response: %w", err)
	}
	return &webhook, nil
}

// Remove a webhook from a repository
func (c *Client) DeleteWebhook(ctx context.Context, owner,
	repo string, webhookID int64) error {
	endpoint := fmt.Sprintf("/repos/%s/%s/hooks/%d", owner, repo, webhookID)

	resp, err := c.doRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("Failed to delete webhook: %w", err)
	}
	defer resp.Body.Close()

	// 204 No Content = success, 404 Not Found = already deleted
	if resp.StatusCode != http.StatusNoContent &&
		resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Failed to delete webhook: %s - %s",
			resp.Status, string(body))
	}

	return nil
}

// Splits "owner/repo" into owner & repo
func ParseRepoFullName(fullName string) (owner, repo string, err error) {
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Invalid repository full name: %s", fullName)
	}
	return parts[0], parts[1], nil
}
