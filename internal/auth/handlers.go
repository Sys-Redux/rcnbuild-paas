package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/Sys-Redux/rcnbuild-paas/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Handlers provides HTTP handlers for authentication
type Handlers struct{}

// Create a new auth handlers instance
func NewHandlers() *Handlers {
	return &Handlers{}
}

// Redirect the user to GitHub OAuth authorization page
func (h *Handlers) HandleGitHubLogin(c *gin.Context) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	redirectURI := os.Getenv("GITHUB_REDIRECT_URI")

	if clientID == "" || redirectURI == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GitHub OAuth not configured",
		})
		return
	}

	// Build GitHub OAuth URL
	// For GitHub Apps, permissions are defined in the app settings
	authURL := "https://github.com/login/oauth/authorize?client_id=" +
		clientID + "&redirect_uri=" + redirectURI

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// Handle the OAuth callback from GitHub
func (h *Handlers) HandleGitHubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing authorization code",
		})
		return
	}

	// Exchange code for access token
	tokenResp, err := exchangeCodeForToken(code)
	if err != nil {
		log.Error().Err(err).Msg("Failed to exchange code for token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to exchange code for token",
		})
		return
	}

	// Fetch user info from GitHub
	githubUser, err := fetchGitHubUser(tokenResp.AccessToken)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch GitHub user")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch GitHub user info",
		})
		return
	}

	// Create or update user in database
	user, err := database.CreateOrUpdateUser(c.Request.Context(),
		githubUser, tokenResp.AccessToken)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create/update user")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create or update user",
		})
		return
	}

	// Generate JWT
	jwtToken, err := GenerateToken(user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate authentication token",
		})
		return
	}

	// Set auth cookie
	SetAuthCookie(c, jwtToken)
	log.Info().
		Str("user_id", user.ID).
		Str("github_username", user.GitHubUsername).
		Msg("User authenticated successfully")

	// Redirect to dashboard
	dashboardURL := os.Getenv("DASHBOARD_URL")
	if dashboardURL == "" {
		dashboardURL = "/dashboard"
	}
	c.Redirect(http.StatusTemporaryRedirect, dashboardURL)
}

// Clear the user's session
func (h *Handlers) HandleLogout(c *gin.Context) {
	ClearAuthCookie(c)
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// Get the current authenticated user
func (h *Handlers) HandleGetMe(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ===========================================
// Internal helpers
// ===========================================

// Represents GitHub's token exchange response
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// Exchange the authorization code for an access token
func exchangeCodeForToken(code string) (*tokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", os.Getenv("GITHUB_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("GITHUB_CLIENT_SECRET"))
	data.Set("code", code)

	req, err := http.NewRequest("POST",
		"https://github.com/login/oauth/access_token",
		bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// Fetch the authenticated user's info from GitHub API
func fetchGitHubUser(accessToken string) (*database.GitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user database.GitHubUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	return &user, nil
}
