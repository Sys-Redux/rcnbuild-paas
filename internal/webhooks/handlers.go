package webhooks

import (
	"io"
	"net/http"

	"github.com/Sys-Redux/rcnbuild-paas/internal/database"
	"github.com/Sys-Redux/rcnbuild-paas/internal/queue"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Provide HTTP handlers for webhooks
type Handlers struct{}

// Create a new webhooks handlers instance
func NewHandlers() *Handlers {
	return &Handlers{}
}

// Handle incoming GitHub webhook
func (h *Handlers) HandleGitHubWebhook(c *gin.Context) {
	// Read the request body for signature verification
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read webhook body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	eventType := c.GetHeader("X-GitHub-Event")
	deliveryID := c.GetHeader("X-GitHub-Delivery")
	signature := c.GetHeader("X-Hub-Signature-256")

	log.Info().
		Str("event", eventType).
		Str("delivery_id", deliveryID).
		Msg("Received GitHub webhook")

	// Only handle push events for now
	if eventType != "push" {
		log.Debug().Str("event", eventType).Msg("Ignoring non-push event")
		c.JSON(http.StatusOK, gin.H{"message": "Event ignored"})
		return
	}

	// Parse push event payload
	pushEvent, err := ParsePushEvent(body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse push event")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid push event"})
		return
	}

	// Find project by repo full name
	project, err := database.GetProjectByRepoFullName(c.Request.Context(),
		pushEvent.Repository.FullName)
	if err != nil {
		log.Warn().
			Str("repo", pushEvent.Repository.FullName).
			Msg("No project found for repository")
		c.JSON(http.StatusOK, gin.H{
			"message": "No associated project found",
		})
		return
	}

	// Validate webhook signature using project's webhook secret
	if project.WebhookSecret == nil || *project.WebhookSecret == "" {
		log.Error().Str("project_id", project.ID).
			Msg("Project has no webhook secret configured")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := ValidateSignature(body, signature,
		*project.WebhookSecret); err != nil {
		log.Warn().
			Err(err).
			Str("project_id", project.ID).
			Msg("Invalid webhook signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Check if this push should deploy
	if !pushEvent.ShouldDeploy() {
		log.Debug().Msg("Push event does not meet deployment criteria")
		c.JSON(http.StatusOK, gin.H{
			"message": "Push event does not trigger deployment",
		})
		return
	}

	// Check if push is to the configured branch
	pushBranch := pushEvent.GetBranch()
	if pushBranch != project.Branch {
		log.Debug().
			Str("push_branch", pushBranch).
			Str("configured_branch", project.Branch).
			Msg("Push to non-configured branch, skipping deployment")
		c.JSON(http.StatusOK, gin.H{
			"message": "Push to non-configured branch, deployment skipped",
			"branch":  pushBranch,
		})
		return
	}

	// Get commit info
	commitSHA, commitMessage, commitAuthor := pushEvent.GetCommitInfo()

	// Create deployment record
	deployment, err := database.CreateDeployment(c.Request.Context(),
		&database.CreateDeploymentInput{
			ProjectID:     project.ID,
			CommitSHA:     commitSHA,
			CommitMessage: &commitMessage,
			CommitAuthor:  &commitAuthor,
			Branch:        &pushBranch,
		})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create deployment record")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create deployment",
		})
		return
	}

	log.Info().
		Str("deployment_id", deployment.ID).
		Str("project_id", project.ID).
		Str("commit", commitSHA[:8]).
		Str("branch", pushBranch).
		Msg("Created deployment record from push event")

	// Enqueue build job w/ Asynq
	_, err = queue.EnqueueBuild(c.Request.Context(), &queue.BuildPayload{
		DeploymentID: deployment.ID,
		ProjectID:    project.ID,
		CommitSHA:    commitSHA,
		Branch:       pushBranch,
		RepoFullName: project.RepoFullName,
		RepoCloneURL: project.RepoURL,
		RootDir:      project.RootDirectory,
		BuildCommand: stringOrDefault(project.BuildCommand, ""),
		StartCommand: stringOrDefault(project.StartCommand, ""),
		Runtime:      stringOrDefault(project.Runtime, ""),
		Port:         project.Port,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to enqueue build job")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to enqueue build job",
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":       "Deployment created",
		"deployment_id": deployment.ID,
		"commit":        commitSHA,
		"branch":        pushBranch,
	})
}

// Helper functions
func stringOrDefault(s *string, def string) string {
	if s != nil {
		return *s
	}
	return def
}
