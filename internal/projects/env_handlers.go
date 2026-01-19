package projects

import (
	"net/http"

	"github.com/Sys-Redux/rcnbuild-paas/internal/auth"
	"github.com/Sys-Redux/rcnbuild-paas/internal/database"
	"github.com/Sys-Redux/rcnbuild-paas/pkg/crypto"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Body for creating/updating an environment variable
type CreateEnvVarRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// List env var for a project
// GET /api/projects/:id/env
func (h *Handlers) HandleListEnvVars(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectID := c.Param("id")
	project, err := database.GetProjectByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	// Check if user has access to the project
	if project.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access Denied"})
		return
	}

	envVars, err := database.GetEnvVarsByProjectID(c.Request.Context(),
		project.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get env vars")
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed to get env vars"})
		return
	}

	// Convert to display format (decrypt values)
	displayVars := database.ToDisplayList(envVars)

	c.JSON(http.StatusOK, gin.H{"env_vars": displayVars})
}

// Create or update an env var for a project
// POST /api/projects/:id/env
func (h *Handlers) HandleCreateEnvVar(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectID := c.Param("id")
	project, err := database.GetProjectByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	// Check if user has access to the project
	if project.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access Denied"})
		return
	}

	var req CreateEnvVarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate key format (only allow alphanumeric and underscores)
	if !isValidEnvKey(req.Key) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key format"})
		return
	}

	// Encrypt the value before storing
	encryptedValue, err := crypto.Encrypt(req.Value)
	if err != nil {
		log.Error().Err(err).Msg("Failed to encrypt env var value")
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed to encrypt env var value"})
		return
	}

	// Create or update the env var in the database
	envVar, err := database.CreateOrUpdateEnvVar(c.Request.Context(),
		project.ID, req.Key, encryptedValue)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create/update env var")
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed to create/update env var"})
		return
	}

	// Return masked display format
	c.JSON(http.StatusCreated, envVar.ToDisplay())
}

// Delete an env var
// DELETE /api/projects/:id/env/:key
func (h *Handlers) HandleDeleteEnvVar(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectID := c.Param("id")
	key := c.Param("key")

	project, err := database.GetProjectByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	// Check if user has access to the project
	if project.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access Denied"})
		return
	}

	if err := database.DeleteEnvVar(c.Request.Context(),
		project.ID, key); err != nil {
		if err.Error() == "env var not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "env var not found"})
			return
		}

		log.Error().Err(err).Msg("Failed to delete env var")
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed to delete env var"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "env var deleted"})
}

// Validate env var key format (alphanumeric and underscores only)
func isValidEnvKey(key string) bool {
	if len(key) == 0 || len(key) > 255 {
		return false
	}

	// Must start w/ letter
	first := key[0]
	if !((first >= 'A' && first <= 'Z') || (first >= 'a' && first <= 'z')) {
		return false
	}

	// Rest can be letters, numbers, or underscores
	for _, c := range key[1:] {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}
