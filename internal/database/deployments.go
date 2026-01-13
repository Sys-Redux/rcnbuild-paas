package database

import (
	"context"
	"errors"
	"time"
)

// Represents state of deployment
type DeploymentStatus string

const (
	DeploymentStatusPending    DeploymentStatus = "pending"
	DeploymentStatusBuilding   DeploymentStatus = "building"
	DeploymentStatusDeploying  DeploymentStatus = "deploying"
	DeploymentStatusLive       DeploymentStatus = "live"
	DeploymentStatusFailed     DeploymentStatus = "failed"
	DeploymentStatusCancelled  DeploymentStatus = "cancelled"
	DeploymentStatusSuperseded DeploymentStatus = "superseded"
)

// Represents a single deployment attempt
type Deployment struct {
	ID            string           `json:"id"`
	ProjectID     string           `json:"project_id"`
	CommitSHA     string           `json:"commit_sha"`
	CommitMessage *string          `json:"commit_message,omitempty"`
	CommitAuthor  *string          `json:"commit_author,omitempty"`
	Branch        *string          `json:"branch,omitempty"`
	Status        DeploymentStatus `json:"status"`
	ImageTag      *string          `json:"image_tag,omitempty"`
	ContainerID   *string          `json:"-"` // Internal use only
	URL           *string          `json:"url,omitempty"`
	BuildLogsURL  *string          `json:"build_logs_url,omitempty"`
	ErrorMessage  *string          `json:"error_message,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	StartedAt     *time.Time       `json:"started_at,omitempty"`
	CompletedAt   *time.Time       `json:"completed_at,omitempty"`
}

// For creating a new deployment
type CreateDeploymentInput struct {
	ProjectID     string
	CommitSHA     string
	CommitMessage *string
	CommitAuthor  *string
	Branch        *string
}

// Creates new deploy w/ status "pending"
func CreateDeployment(ctx context.Context,
	input *CreateDeploymentInput) (*Deployment, error) {
	query := `
		INSERT INTO deployments (
			project_id, commit_sha, commit_message, commit_author,
			branch, status
		) VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING id, project_id, commit_sha, commit_message, commit_author,
			branch, status, image_tag, container_id, url, build_logs_url,
			error_message, created_at, started_at, completed_at
	`

	var d Deployment
	err := pool.QueryRow(ctx, query,
		input.ProjectID,
		input.CommitSHA,
		input.CommitMessage,
		input.CommitAuthor,
		input.Branch,
	).Scan(
		&d.ID, &d.ProjectID, &d.CommitSHA, &d.CommitMessage, &d.CommitAuthor,
		&d.Branch, &d.Status, &d.ImageTag, &d.ContainerID, &d.URL,
		&d.BuildLogsURL, &d.ErrorMessage, &d.CreatedAt, &d.StartedAt,
		&d.CompletedAt,
	)

	if err != nil {
		return nil, err
	}
	return &d, nil
}

// Retrieve deployment by ID
func GetDeploymentByID(ctx context.Context, id string) (*Deployment, error) {
	query := `
		SELECT id, project_id, commit_sha, commit_message, commit_author,
			branch, status, image_tag, container_id, url, build_logs_url,
			error_message, created_at, started_at, completed_at
		FROM deployments
		WHERE id = $1
	`

	var d Deployment
	err := pool.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.ProjectID, &d.CommitSHA, &d.CommitMessage, &d.CommitAuthor,
		&d.Branch, &d.Status, &d.ImageTag, &d.ContainerID, &d.URL,
		&d.BuildLogsURL, &d.ErrorMessage, &d.CreatedAt, &d.StartedAt,
		&d.CompletedAt,
	)

	if err != nil {
		return nil, err
	}
	return &d, nil
}

// Return deploys for a project
func GetDeploymentsByProjectID(ctx context.Context,
	projectID string, limit int) ([]*Deployment, error) {
	query := `
		SELECT id, project_id, commit_sha, commit_message, commit_author,
			branch, status, image_tag, container_id, url, build_logs_url,
			error_message, created_at, started_at, completed_at
		FROM deployments
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := pool.Query(ctx, query, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*Deployment
	for rows.Next() {
		var d Deployment
		err := rows.Scan(
			&d.ID, &d.ProjectID, &d.CommitSHA, &d.CommitMessage, &d.CommitAuthor,
			&d.Branch, &d.Status, &d.ImageTag, &d.ContainerID, &d.URL,
			&d.BuildLogsURL, &d.ErrorMessage, &d.CreatedAt, &d.StartedAt,
			&d.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		deployments = append(deployments, &d)
	}
	return deployments, nil
}

// Returns current live deployment
func GetLiveDeployment(ctx context.Context,
	projectID string) (*Deployment, error) {
	query := `
		SELECT id, project_id, commit_sha, commit_message, commit_author,
			branch, status, image_tag, container_id, url, build_logs_url,
			error_message, created_at, started_at, completed_at
		FROM deployments
		WHERE project_id = $1 AND status = 'live'
		LIMIT 1
	`

	var d Deployment
	err := pool.QueryRow(ctx, query, projectID).Scan(
		&d.ID, &d.ProjectID, &d.CommitSHA, &d.CommitMessage, &d.CommitAuthor,
		&d.Branch, &d.Status, &d.ImageTag, &d.ContainerID, &d.URL,
		&d.BuildLogsURL, &d.ErrorMessage, &d.CreatedAt, &d.StartedAt,
		&d.CompletedAt,
	)

	if err != nil {
		return nil, err
	}
	return &d, nil
}

// Updates status & optionally sets error message
func UpdateDeploymentStatus(ctx context.Context, id string,
	status DeploymentStatus, errorMsg *string) error {
	query := `
		UPDATE deployments
		SET status = $2, error_message = $3, updated_at = NOW()
		WHERE id = $1
	`

	result, err := pool.Exec(ctx, query, id, status, errorMsg)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("deployment not found")
	}

	return nil
}

// Marks deployment as building
func StartDeploymentBuild(ctx context.Context, id string) error {
	query := `
		UPDATE deployments
		SET status = 'building', started_at = NOW()
		WHERE id = $1
	`

	result, err := pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("deployment not found")
	}

	return nil
}

// Marks build complete & stores image tag
func SetDeploymentBuilt(ctx context.Context, id string,
	imageTag string) error {
	query := `
		UPDATE deployments
		SET status = 'deploying', image_tag = $2
		WHERE id = $1
	`

	result, err := pool.Exec(ctx, query, id, imageTag)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("deployment not found")
	}

	return nil
}

// Marks deployment as live & stores container info
func SetDeploymentLive(ctx context.Context, id string,
	containerID string, url string) error {
	query := `
		UPDATE deployments
		SET status = 'live', container_id = $2, url = $3,
			completed_at = NOW()
		WHERE id = $1
	`

	result, err := pool.Exec(ctx, query, id, containerID, url)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("deployment not found")
	}

	return nil
}

// Marks all other 'live' deployments for a project as 'superseded'
func SupersededOldDeployments(ctx context.Context, projectID string,
	excludeDeploymentID string) error {
	query := `
		UPDATE deployments
		SET status = 'superseded', completed_at = NOW()
		WHERE project_id = $1 AND status = 'live' AND id != $2
	`

	_, err := pool.Exec(ctx, query, projectID, excludeDeploymentID)
	return err
}

// Marks deployment as failed
func SetDeploymentFailed(ctx context.Context, id string,
	errorMsg string) error {
	query := `
		UPDATE deployments
		SET status = 'failed', error_message = $2, completed_at = NOW()
		WHERE id = $1
	`

	result, err := pool.Exec(ctx, query, id, errorMsg)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("deployment not found")
	}

	return nil
}

// Marks deployment as cancelled
func CancelDeployment(ctx context.Context, id string) error {
	query := `
		UPDATE deployments
		SET status = 'cancelled', completed_at = NOW()
		WHERE id = $1 AND status IN ('pending', 'building', 'deploying')
	`

	result, err := pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("deployment not found or cannot be cancelled")
	}

	return nil
}

// Removes deployment record (cleanup)
func DeleteDeployment(ctx context.Context, id string) error {
	query := `DELETE FROM deployments WHERE id = $1`

	result, err := pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("deployment not found")
	}

	return nil
}

// Removes all deployments for a project (when deleting a project)
func DeleteDeploymentsByProjectID(ctx context.Context,
	projectID string) error {
	query := `DELETE FROM deployments WHERE project_id = $1`

	_, err := pool.Exec(ctx, query, projectID)
	return err
}
