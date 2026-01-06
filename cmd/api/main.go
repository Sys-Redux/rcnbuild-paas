package api

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Set up Zero Logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Create Gin router
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
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
		webhook := api.Group("/webhook")
		{
			webhook.POST("/github", handleGitHubWebhook)
		}
	}

	// Get port from environment
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
}
