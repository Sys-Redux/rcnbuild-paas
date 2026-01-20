package queue

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TypeBuildProject  = "build:project"
	TypeDeployProject = "deploy:project"
)

// Data for build job
type BuildPayload struct {
	DeploymentID string `json:"deployment_id"`
	ProjectID    string `json:"project_id"`
	CommitSHA    string `json:"commit_sha"`
	Branch       string `json:"branch"`
	RepoFullName string `json:"repo_full_name"`
	RepoCloneURL string `json:"repo_clone_url"`
	RootDir      string `json:"root_dir"`
	BuildCommand string `json:"build_command"`
	StartCommand string `json:"start_command"`
	Runtime      string `json:"runtime"`
	Port         int    `json:"port"`
}

// Data for deploy job
type DeployPayload struct {
	DeploymentID string `json:"deployment_id"`
	ProjectID    string `json:"project_id"`
	ProjectSlug  string `json:"project_slug"`
	ImageTag     string `json:"image_tag"`
	Port         int    `json:"port"`
}

// Create new build task
func NewBuildTask(payload *BuildPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeBuildProject, data,
		asynq.MaxRetry(3),
		asynq.Timeout(30*time.Minute),
		asynq.Queue("builds"),
	), nil
}

// Create new deploy task
func NewDeployTask(payload *DeployPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeDeployProject, data,
		asynq.MaxRetry(3),
		asynq.Timeout(5*time.Minute),
		asynq.Queue("deployments"),
	), nil
}
