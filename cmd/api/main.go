package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sys-Redux/rcnbuild-paas/internal/auth"
	"github.com/Sys-Redux/rcnbuild-paas/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	// Setup zerolog with pretty console output
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer database.Close()

	// Set Gin mode based on environment
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "rcnbuild-api",
		})
	})

	// API Routes
	api := r.Group("/api")
	{
		// Auth routes
		authGroup := api.Group("/auth")
		{
			authGroup.GET("/github", handleGitHubLogin)
			authGroup.GET("/github/callback", handleGitHubCallback)
			authGroup.POST("/logout", handleLogout)
			authGroup.GET("/me", auth.AuthRequired(), handleGetMe)
		}

		// Webhook routes
		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/github", handleGitHubWebhook)
		}
	}

	// Get configuration from environment
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	host := os.Getenv("API_HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	addr := fmt.Sprintf("%s:%s", host, port)

	// Create HTTP server
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info().Str("addr", addr).Msg("Starting RCNbuild API server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// Give outstanding requests 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}

// ===========================================
// Auth Handlers
// ===========================================

// handleGitHubLogin redirects the user to GitHub App OAuth authorization page
func handleGitHubLogin(c *gin.Context) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	redirectURI := os.Getenv("GITHUB_REDIRECT_URI")

	if clientID == "" || redirectURI == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GitHub OAuth not configured",
		})
		return
	}

	// Build GitHub App OAuth URL
	// For GitHub Apps permissions are defined in the app settings
	// No scopes needed - the app's permissions are pre-configured
	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s",
		clientID,
		redirectURI,
	)

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// handleGitHubCallback handles the OAuth callback from GitHub
func handleGitHubCallback(c *gin.Context) {
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
	jwtToken, err := auth.GenerateToken(user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate authentication token",
		})
		return
	}

	// Set auth cookie
	auth.SetAuthCookie(c, jwtToken)
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

// TokenResponse represents GitHub's token exchange response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// Exchanges authorization code for access token
func exchangeCodeForToken(code string) (*TokenResponse, error) {
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

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// Fetch authenticated user's info from GitHub API
func fetchGitHubUser(accessToken string) (*database.GitHubUser, error) {
	req, err := http.NewRequest("GET",
		"https://api.github.com/user", nil)
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

// handleLogout clears the user's session
func handleLogout(c *gin.Context) {
	auth.ClearAuthCookie(c)
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// handleGetMe returns the current authenticated user
func handleGetMe(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ===========================================
// Webhook Handlers (TODO: Implement)
// ===========================================

// handleGitHubWebhook processes incoming GitHub webhooks
func handleGitHubWebhook(c *gin.Context) {
	// TODO: Validate webhook signature using GITHUB_WEBHOOK_SECRET
	// TODO: Parse webhook payload
	// TODO: Handle push events to trigger deployments

	eventType := c.GetHeader("X-GitHub-Event")
	deliveryID := c.GetHeader("X-GitHub-Delivery")

	log.Info().
		Str("event", eventType).
		Str("delivery_id", deliveryID).
		Msg("Received GitHub webhook")

	c.JSON(http.StatusOK, gin.H{
		"message":     "Webhook received",
		"event":       eventType,
		"delivery_id": deliveryID,
	})
}
