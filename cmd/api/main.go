package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		auth := api.Group("/auth")
		{
			auth.GET("/github", handleGitHubLogin)
			auth.GET("/github/callback", handleGitHubCallback)
			auth.POST("/logout", handleLogout)
			auth.GET("/me", handleGetMe)
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
// Auth Handlers (TODO: Implement)
// ===========================================

// handleGitHubLogin redirects the user to GitHub OAuth authorization page
func handleGitHubLogin(c *gin.Context) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	redirectURI := os.Getenv("GITHUB_REDIRECT_URI")

	if clientID == "" || redirectURI == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GitHub OAuth not configured",
		})
		return
	}

	// Build GitHub OAuth URL
	// Scopes: read:user (profile), user:email (email), repo (access repos)
	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=read:user,user:email,repo",
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

	// TODO: Exchange code for access token
	// TODO: Fetch user info from GitHub
	// TODO: Create/update user in database
	// TODO: Generate JWT and set cookie
	// TODO: Redirect to dashboard

	c.JSON(http.StatusOK, gin.H{
		"message": "GitHub callback received",
		"code":    code,
		"todo":    "Exchange code for token and create session",
	})
}

// handleLogout clears the user's session
func handleLogout(c *gin.Context) {
	// TODO: Clear JWT cookie
	// TODO: Invalidate session in Redis (optional)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// handleGetMe returns the current authenticated user
func handleGetMe(c *gin.Context) {
	// TODO: Extract JWT from cookie/header
	// TODO: Validate JWT
	// TODO: Return user info from database

	c.JSON(http.StatusUnauthorized, gin.H{
		"error": "Not authenticated",
	})
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
