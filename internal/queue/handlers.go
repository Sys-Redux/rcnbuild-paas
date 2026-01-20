package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Sys-Redux/rcnbuild-paas/internal/builds"
	"github.com/Sys-Redux/rcnbuild-paas/internal/containers"
	"github.com/Sys-Redux/rcnbuild-paas/internal/database"
	"github.com/Sys-Redux/rcnbuild-paas/pkg/crypto"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

// Process build jobs
func HandleBuildTask(ctx context.Context, t *asynq.Task) error {
	var payload BuildPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal build payload: %w", err)
	}

	log.Info().
		Str("deployment_id", payload.DeploymentID).
		Str("commit", payload.CommitSHA[:8]).
		Msg("Started processing build job")

	// Update deployment status
	if err := database.StartDeploymentBuild(ctx,
		payload.DeploymentID); err != nil {
		return fmt.Errorf("failed to start deployment build: %w", err)
	}

	// Create build directory (temporary)
	buildDir, err := os.MkdirTemp("", "recnbuild-*")
	if err != nil {
		return failBuild(ctx, payload.DeploymentID,
			"failed to create build directory", err)
	}
	defer os.RemoveAll(buildDir) // Cleanup after build

	// Clone repo
	log.Info().Str("repo", payload.RepoFullName).Msg("Cloning repository")
	if err := cloneRepo(ctx, payload.RepoCloneURL, payload.CommitSHA,
		buildDir); err != nil {
		return failBuild(ctx, payload.DeploymentID,
			"failed to clone repository", err)
	}

	// Determine working directory
	workDir := buildDir
	if payload.RootDir != "" && payload.RootDir != "." {
		workDir = filepath.Join(buildDir, payload.RootDir)
	}

	// Make Dockerfile if it doesn't exist
	dockerfilePath := filepath.Join(workDir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		log.Info().Str("runtime", payload.Runtime).Msg("Generating Dockerfile")
		runtimeInfo := &builds.RuntimeInfo{
			Runtime:      builds.Runtime(payload.Runtime),
			BuildCommand: payload.BuildCommand,
			StartCommand: payload.StartCommand,
			Port:         payload.Port,
		}
		dockerfile := builds.GetDockerfileForRuntime(runtimeInfo,
			payload.BuildCommand, payload.StartCommand)
		if err := os.WriteFile(dockerfilePath, []byte(dockerfile),
			0644); err != nil {
			return failBuild(ctx, payload.DeploymentID,
				"failed to write Dockerfile", err)
		}
	}

	// Build container image
	registryURL := os.Getenv("REGISTRY_URL")
	if registryURL == "" {
		registryURL = "localhost:5000"
	}
	imageTag := fmt.Sprintf("%s/%s:%s", registryURL,
		payload.ProjectID, payload.CommitSHA[:8])
	log.Info().Str("image", imageTag).Msg("Building container image")
	if err := buildImage(ctx, workDir, imageTag); err != nil {
		return failBuild(ctx, payload.DeploymentID,
			"failed to build container image", err)
	}

	// Push to docker registry
	log.Info().Str("image", imageTag).Msg("Pushing to registry")
	if err := pushImage(ctx, imageTag); err != nil {
		return failBuild(ctx, payload.DeploymentID,
			"failed to push container image", err)
	}

	// Update w/ image tag
	if err := database.SetDeploymentBuilt(ctx, payload.DeploymentID,
		imageTag); err != nil {
		return fmt.Errorf("failed to set deployment built: %w", err)
	}

	log.Info().
		Str("deployment_id", payload.DeploymentID).
		Str("image", imageTag).
		Msg("Build completed successfully")

	// Get project for deploy info
	project, err := database.GetProjectByID(ctx, payload.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Enqueue deploy job
	_, err = EnqueueDeploy(ctx, &DeployPayload{
		DeploymentID: payload.DeploymentID,
		ProjectID:    payload.ProjectID,
		ProjectSlug:  project.Slug,
		ImageTag:     imageTag,
		Port:         payload.Port,
	})
	if err != nil {
		return fmt.Errorf("failed to enqueue deploy job: %w", err)
	}

	return nil
}

// Process deploy jobs
func HandleDeployTask(ctx context.Context, t *asynq.Task) error {
	var payload DeployPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal deploy payload: %w", err)
	}

	log.Info().
		Str("deployment_id", payload.DeploymentID).
		Str("image", payload.ImageTag).
		Msg("Starting deployment")

	// Update deployment status
	if err := database.UpdateDeploymentStatus(ctx, payload.DeploymentID,
		database.DeploymentStatusDeploying, nil); err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	// Fetch env vars for the project
	envVars, err := database.GetEnvVarsAsMap(ctx, payload.ProjectID,
		crypto.Decrypt)
	if err != nil {
		return failDeploy(ctx, payload.DeploymentID,
			"failed to fetch environment variables", err)
	}

	// add PORT to env
	envVars["PORT"] = fmt.Sprintf("%d", payload.Port)

	// Deploy container
	baseDomain := os.Getenv("BASE_DOMAIN")
	if baseDomain == "" {
		baseDomain = "rcnbuild.dev"
	}

	containerID, err := containers.Deploy(ctx, &containers.DeployConfig{
		ContainerName: fmt.Sprintf("rcn-%s", payload.ProjectSlug),
		ImageTag:      payload.ImageTag,
		Port:          payload.Port,
		EnvVars:       envVars,
		Slug:          payload.ProjectSlug,
		BaseDomain:    baseDomain,
	})
	if err != nil {
		return failDeploy(ctx, payload.DeploymentID,
			"failed to deploy container", err)
	}

	// Mark old deployments superseded
	if err := database.SupersededOldDeployments(ctx, payload.ProjectID,
		payload.DeploymentID); err != nil {
		return fmt.Errorf("failed to supersede old deployments: %w", err)
	}

	// Update deployment as live
	deployURL := fmt.Sprintf("https://%s.%s", payload.ProjectSlug, baseDomain)
	if err := database.SetDeploymentLive(ctx, payload.DeploymentID,
		containerID, deployURL); err != nil {
		return fmt.Errorf("failed to set deployment deployed: %w", err)
	}

	log.Info().
		Str("deployment_id", payload.DeploymentID).
		Str("container_id", containerID).
		Str("url", deployURL).
		Msg("Deployment completed successfully")

	return nil
}

// Helper functions
// Clone repo
func cloneRepo(ctx context.Context, cloneURL, commitSHA,
	destDir string) error {
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1",
		cloneURL, destDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %s, %w", string(output), err)
	}

	// Fetch specific commit if not HEAD
	fetchCmd := exec.CommandContext(ctx, "git", "-C", destDir,
		"fetch", "origin", commitSHA)
	// Ignore error if commit is HEAD
	fetchCmd.CombinedOutput()

	// Checkout specific commit
	checkoutCmd := exec.CommandContext(ctx, "git", "-C", destDir,
		"checkout", commitSHA)
	if output, err := checkoutCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout failed: %s, %w", string(output), err)
	}

	return nil
}

// Build container image using Docker CLI
func buildImage(ctx context.Context, workDir, imageTag string) error {
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", imageTag, ".")
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker build failed: %s, %w", string(output), err)
	}
	return nil
}

// Push docker image
func pushImage(ctx context.Context, imageTag string) error {
	cmd := exec.CommandContext(ctx, "docker", "push", imageTag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker push failed: %s, %w", string(output), err)
	}
	return nil
}

// Fail build helper
func failBuild(ctx context.Context, deploymentID,
	message string, err error) error {
	fullMessage := fmt.Sprintf("%s: %v", message, err)
	log.Error().Err(err).Str("deployment_id", deploymentID).Msg(message)
	database.SetDeploymentFailed(ctx, deploymentID, fullMessage)
	return fmt.Errorf(fullMessage)
}

// Fail deploy helper
func failDeploy(ctx context.Context, deploymentID,
	message string, err error) error {
	fullMessage := fmt.Sprintf("%s: %v", message, err)
	log.Error().Err(err).Str("deployment_id", deploymentID).Msg(message)
	database.SetDeploymentFailed(ctx, deploymentID, fullMessage)
	return fmt.Errorf(fullMessage)
}
