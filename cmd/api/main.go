package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sys-Redux/rcnbuild-paas/internal/auth"
	"github.com/Sys-Redux/rcnbuild-paas/internal/database"
	"github.com/Sys-Redux/rcnbuild-paas/internal/projects"
	"github.com/Sys-Redux/rcnbuild-paas/internal/webhooks"
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

	// Initialize handlers
	authHandlers := auth.NewHandlers()
	projectHandlers := projects.NewHandlers()
	webhookHandlers := webhooks.NewHandlers()

	// API Routes
	api := r.Group("/api")
	{
		// Auth routes
		authGroup := api.Group("/auth")
		{
			authGroup.GET("/github", authHandlers.HandleGitHubLogin)
			authGroup.GET("/github/callback", authHandlers.HandleGitHubCallback)
			authGroup.POST("/logout", authHandlers.HandleLogout)
			authGroup.GET("/me", auth.AuthRequired(), authHandlers.HandleGetMe)
		}

		// GitHub repos (for selecting repo to deploy)
		api.GET("/repos", auth.AuthRequired(),
			projectHandlers.HandleListRepos)

		// Project routes
		projectsGroup := api.Group("/projects")
		projectsGroup.Use(auth.AuthRequired())
		{
			projectsGroup.GET("", projectHandlers.HandleListProjects)
			projectsGroup.POST("", projectHandlers.HandleCreateProject)
			projectsGroup.GET("/:id", projectHandlers.HandleGetProject)
			projectsGroup.PATCH("/:id", projectHandlers.HandleUpdateProject)
			projectsGroup.DELETE("/:id", projectHandlers.HandleDeleteProject)

			// Environment variable routes
			projectsGroup.GET("/:id/env", projectHandlers.HandleListEnvVars)
			projectsGroup.POST("/:id/env", projectHandlers.HandleCreateEnvVar)
			projectsGroup.DELETE("/:id/env/:key",
				projectHandlers.HandleDeleteEnvVar)
		}

		// Webhook routes (no auth - handled via secret)
		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/github", webhookHandlers.HandleGitHubWebhook)
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
