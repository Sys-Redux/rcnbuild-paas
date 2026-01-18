package projects

import (
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/Sys-Redux/rcnbuild-paas/internal/auth"
	"github.com/Sys-Redux/rcnbuild-paas/internal/builds"
	"github.com/Sys-Redux/rcnbuild-paas/internal/database"
	"github.com/Sys-Redux/rcnbuild-paas/internal/github"
	"github.com/Sys-Redux/rcnbuild-paas/pkg/crypto"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Holds dependencies for project handlers
type Handlers struct{}

// Create Handlers instance
func NewHandlers() *Handlers {
	return &Handlers{}
}

// Query parms for listing repos
type ListReposRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// Body for creating a new project
type CreateProjectRequest struct {
	RepoFullName  string  `json:"repo_full_name" binding:"required"`
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	Branch        string  `json:"branch"`
	RootDirectory string  `json:"root_directory"`
	BuildCommand  *string `json:"build_command"`
	StartCommand  *string `json:"start_command"`
	Port          int     `json:"port"`
}

// Body for updating a project
type UpdateProjectRequest struct {
	Name          *string `json:"name"`
	Branch        *string `json:"branch"`
	RootDirectory *string `json:"root_directory"`
	BuildCommand  *string `json:"build_command"`
	StartCommand  *string `json:"start_command"`
	Port          *int    `json:"port"`
}

// Lists repos the user can deploy
// GET /api/repos
func (h *Handlers) HandleListRepos(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req ListReposRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	// Get user's access token
	accessToken, err := database.GetUserAccessToken(c.Request.Context(),
		user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user access token")
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed to get user access token"})
		return
	}

	// Create GitHub client & list repos
	ghClient := github.NewClient(accessToken)
	repos, err := ghClient.ListUserRepos(c.Request.Context(),
		req.Page, req.PageSize)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list user repos")
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed to list user repos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"repos": repos,
		"page":  req.Page,
	})
}

// Lists user's projects
// GET /api/projects
func (h *Handlers) HandleListProjects(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projects, err := database.GetProjectsByUserID(c.Request.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user projects")
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed to get user projects"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
	})
}

// Create a new project from a github repo
// POST /api/projects
func (h *Handlers) HandleCreateProject(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	// Parse repo full name
	owner, repoName, err := github.ParseRepoFullName(req.RepoFullName)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid repo full name"})
		return
	}

	// Get user's access token
	accessToken, err := database.GetUserAccessToken(c.Request.Context(),
		user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user access token")
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed to get user access token"})
		return
	}

	// Create github client
	ghClient := github.NewClient(accessToken)

	// Verify repo exists & user has permissions
	repo, err := ghClient.GetRepo(c.Request.Context(), owner, repoName)
	if err != nil {
		log.Error().Err(err).Str("repo", req.RepoFullName).Msg(
			"Failed to get github repo")
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "failed to access github repo"})
		return
	}

	// Check if exists already
	existing, _ := database.GetProjectByRepoFullName(c.Request.Context(),
		req.RepoFullName)
	if existing != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "project for this repo already exists"})
		return
	}

	// Set defaults
	projectName := req.Name
	if projectName == "" {
		projectName = repo.Name
	}

	branch := req.Branch
	if branch == "" {
		branch = repo.DefaultBranch
	}

	rootDir := req.RootDirectory
	if rootDir == "" {
		rootDir = "."
	}

	// Generate slug
	slug := req.Slug
	if slug == "" {
		slug = generateSlug(projectName)
	}

	// Ensure slug is unique
	for {
		exists, _ := database.SlugExists(c.Request.Context(), slug)
		if !exists {
			break
		}
		slug = slug + "-" + randomSuffix()
	}

	// Detect runtime
	runtimeInfo, err := builds.DetectRuntime(c.Request.Context(),
		ghClient, owner, repoName, branch, rootDir)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to detect runtime, using defaults")
		runtimeInfo = &builds.RuntimeInfo{
			Runtime: builds.RuntimeUnknown,
			Port:    3000,
		}
	}

	// Use detected values or user-provided overrides
	buildCmd := req.BuildCommand
	if buildCmd == nil && runtimeInfo.BuildCommand != "" {
		buildCmd = &runtimeInfo.BuildCommand
	}

	startCmd := req.StartCommand
	if startCmd == nil && runtimeInfo.StartCommand != "" {
		startCmd = &runtimeInfo.StartCommand
	}

	port := req.Port
	if port == 0 {
		port = runtimeInfo.Port
	}

	runtime := string(runtimeInfo.Runtime)

	// Generate webhook secret
	webhookSecret, err := github.GenerateWebhookSecret()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate webhook secret")
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed to generate webhook secret"})
		return
	}

	// Create webhook on github
	webhookURL := os.Getenv("API_URL") + "/api/webhooks/github"
	webhook, err := ghClient.CreateWebhook(c.Request.Context(),
		owner, repoName, webhookURL, webhookSecret)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create github webhook")
		// Continue anyway, webhook can be created later
	}

	// Create project in database
	input := &database.CreateProjectInput{
		UserId:        user.ID,
		Name:          projectName,
		Slug:          slug,
		RepoFullName:  req.RepoFullName,
		RepoURL:       repo.HTMLURL,
		Branch:        branch,
		RootDirectory: rootDir,
		BuildCommand:  buildCmd,
		StartCommand:  startCmd,
		Runtime:       &runtime,
		Port:          port,
	}

	project, err := database.CreateProject(c.Request.Context(), input)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create project in database")
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed to create project"})
		return
	}

	// Store webhook info
	if webhook != nil {
		// Encrypt webhook secret
		encryptedSecret, err := crypto.Encrypt(webhookSecret)
		if err != nil {
			log.Error().Err(err).Msg("Failed to encrypt webhook secret")
		} else {
			if err := database.SetProjectWebhook(c.Request.Context(),
				project.ID, webhook.ID, encryptedSecret); err != nil {
				log.Error().Err(err).Msg("Failed to store webhook info")
			}
		}
	}

	log.Info().
		Str("project_id", project.ID).
		Str("repo", req.RepoFullName).
		Str("runtime", runtime).
		Msg("Created new project")

	c.JSON(http.StatusCreated, gin.H{
		"project":      project,
		"runtime_info": runtimeInfo,
	})
}

// Returns a specific project
// GET /api/projects/:id
func (h *Handlers) HandleGetProject(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	projectID := c.Param("id")
	project, err := database.GetProjectByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Verify ownership
	if project.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Get latest deployment
	deployment, _ := database.GetLiveDeployment(c.Request.Context(), projectID)

	c.JSON(http.StatusOK, gin.H{
		"project":           project,
		"latest_deployment": deployment,
	})
}

// Update project settings
// PATCH /api/projects/:id
func (h *Handlers) HandleUpdateProject(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	projectID := c.Param("id")
	project, err := database.GetProjectByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Verify ownership
	if project.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build update input
	updateInput := &database.UpdateProjectInput{
		Name:          req.Name,
		Branch:        req.Branch,
		RootDirectory: req.RootDirectory,
		BuildCommand:  req.BuildCommand,
		StartCommand:  req.StartCommand,
		Port:          req.Port,
	}

	updatedProject, err := database.UpdateProject(c.Request.Context(), projectID, updateInput)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update project")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, updatedProject)
}

// Delete a project and its resources
// DELETE /api/projects/:id
func (h *Handlers) HandleDeleteProject(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	projectID := c.Param("id")
	project, err := database.GetProjectByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Verify ownership
	if project.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Delete webhook from GitHub if it exists
	if project.WebhookID != nil {
		accessToken, err := database.GetUserAccessToken(c.Request.Context(), user.ID)
		if err == nil {
			owner, repoName, err := github.ParseRepoFullName(project.RepoFullName)
			if err == nil {
				ghClient := github.NewClient(accessToken)
				if err := ghClient.DeleteWebhook(c.Request.Context(), owner, repoName, *project.WebhookID); err != nil {
					log.Warn().Err(err).Msg("Failed to delete GitHub webhook")
				}
			}
		}
	}

	// Delete all deployments
	if err := database.DeleteDeploymentsByProjectID(c.Request.Context(), projectID); err != nil {
		log.Error().Err(err).Msg("Failed to delete deployments")
	}

	// Delete all env vars
	if err := database.DeleteAllEnvVar(c.Request.Context(), projectID); err != nil {
		log.Error().Err(err).Msg("Failed to delete env vars")
	}

	// Delete the project
	if err := database.DeleteProject(c.Request.Context(), projectID); err != nil {
		log.Error().Err(err).Msg("Failed to delete project")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	log.Info().
		Str("project_id", projectID).
		Str("repo", project.RepoFullName).
		Msg("Project deleted successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Project deleted successfully",
	})
}

// Creates a URL-safe slug from a project name
func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces and underscores with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// Remove any character that isn't alphanumeric or hyphen
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")

	// Remove consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	// Limit length
	if len(slug) > 50 {
		slug = slug[:50]
	}

	return slug
}

// randomSuffix generates a short random suffix for slugs
func randomSuffix() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 4)
	for i := range result {
		result[i] = chars[i%len(chars)] // Simple, not crypto-random
	}
	return string(result)
}
