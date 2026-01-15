package database

import (
	"context"
	"errors"
	"time"
)

// Project represents a deployed application
type Project struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	RepoFullName  string    `json:"repo_full_name"`
	RepoURL       string    `json:"repo_url"`
	Branch        string    `json:"branch"`
	RootDirectory string    `json:"root_directory"`
	BuildCommand  *string   `json:"build_command,omitempty"`
	StartCommand  *string   `json:"start_command,omitempty"`
	Runtime       *string   `json:"runtime,omitempty"`
	Port          int       `json:"port"`
	WebhookID     *int64    `json:"-"`
	WebhookSecret *string   `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// For creating a new project
type CreateProjectInput struct {
	UserId        string
	Name          string
	Slug          string
	RepoFullName  string
	RepoURL       string
	Branch        string
	RootDirectory string
	BuildCommand  *string
	StartCommand  *string
	Runtime       *string
	Port          int
}

// Contains fields that can be updated
type UpdateProjectInput struct {
	Name          *string
	Branch        *string
	RootDirectory *string
	BuildCommand  *string
	StartCommand  *string
	Runtime       *string
	Port          *int
}

// Inserts a new project in database
func CreateProject(ctx context.Context,
	input *CreateProjectInput) (*Project, error) {
	query := `
		INSERT INTO projects (
			user_id, name, slug, repo_full_name, repo_url,
			branch, root_directory, build_command, start_command,
			runtime, port
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, user_id, name, slug, repo_full_name, repo_url,
			branch, root_directory, build_command, start_command,
			runtime, port, webhook_id, webhook_secret, created_at, updated_at
	`

	var p Project
	err := pool.QueryRow(ctx, query,
		input.UserId,
		input.Name,
		input.Slug,
		input.RepoFullName,
		input.RepoURL,
		input.Branch,
		input.RootDirectory,
		input.BuildCommand,
		input.StartCommand,
		input.Runtime,
		input.Port,
	).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Slug, &p.RepoFullName, &p.RepoURL,
		&p.Branch, &p.RootDirectory, &p.BuildCommand, &p.StartCommand,
		&p.Runtime, &p.Port, &p.WebhookID, &p.WebhookSecret,
		&p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Retrieves project by its UUID
func GetProjectByID(ctx context.Context, id string) (*Project, error) {
	query := `
		SELECT
			id, user_id, name, slug, repo_full_name, repo_url,
			branch, root_directory, build_command, start_command,
			runtime, port, webhook_id, webhook_secret,
			created_at, updated_at
		FROM projects
		WHERE id = $1
	`

	var p Project
	err := pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Slug, &p.RepoFullName, &p.RepoURL,
		&p.Branch, &p.RootDirectory, &p.BuildCommand, &p.StartCommand,
		&p.Runtime, &p.Port, &p.WebhookID, &p.WebhookSecret,
		&p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Retrieves project by its slug
func GetProjectBySlug(ctx context.Context, slug string) (*Project, error) {
	query := `
		SELECT
			id, user_id, name, slug, repo_full_name, repo_url,
			branch, root_directory, build_command, start_command,
			runtime, port, webhook_id, webhook_secret,
			created_at, updated_at
		FROM projects
		WHERE slug = $1
	`

	var p Project
	err := pool.QueryRow(ctx, query, slug).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Slug, &p.RepoFullName, &p.RepoURL,
		&p.Branch, &p.RootDirectory, &p.BuildCommand, &p.StartCommand,
		&p.Runtime, &p.Port, &p.WebhookID, &p.WebhookSecret,
		&p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Gets project by repo full name
func GetProjectByRepoFullName(ctx context.Context,
	repoFullName string) (*Project, error) {
	query := `
		SELECT
			id, user_id, name, slug, repo_full_name, repo_url,
			branch, root_directory, build_command, start_command,
			runtime, port, webhook_id, webhook_secret,
			created_at, updated_at
		FROM projects
		WHERE repo_full_name = $1
	`

	var p Project
	err := pool.QueryRow(ctx, query, repoFullName).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Slug, &p.RepoFullName, &p.RepoURL,
		&p.Branch, &p.RootDirectory, &p.BuildCommand, &p.StartCommand,
		&p.Runtime, &p.Port, &p.WebhookID, &p.WebhookSecret,
		&p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Get projects owned by a user
func GetProjectsByUserID(ctx context.Context,
	userID string) ([]*Project, error) {
	query := `
		SELECT
			id, user_id, name, slug, repo_full_name, repo_url,
			branch, root_directory, build_command, start_command,
			runtime, port, webhook_id, webhook_secret,
			created_at, updated_at
		FROM projects
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		var p Project
		err := rows.Scan(
			&p.ID, &p.UserID, &p.Name, &p.Slug, &p.RepoFullName, &p.RepoURL,
			&p.Branch, &p.RootDirectory, &p.BuildCommand, &p.StartCommand,
			&p.Runtime, &p.Port, &p.WebhookID, &p.WebhookSecret,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, &p)
	}

	return projects, nil
}

// Update a projects settings
func UpdateProject(ctx context.Context, id string,
	input *UpdateProjectInput) (*Project, error) {
	query := `
		UPDATE projects SET
			name = COALESCE($2, name),
			branch = COALESCE($3, branch),
			root_directory = COALESCE($4, root_directory),
			build_command = COALESCE($5, build_command),
			start_command = COALESCE($6, start_command),
			runtime = COALESCE($7, runtime),
			port = COALESCE($8, port),
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id, user_id, name, slug, repo_full_name, repo_url,
			branch, root_directory, build_command, start_command,
			runtime, port, webhook_id, webhook_secret,
			created_at, updated_at
	`

	var p Project
	err := pool.QueryRow(ctx, query,
		id,
		input.Name,
		input.Branch,
		input.RootDirectory,
		input.BuildCommand,
		input.StartCommand,
		input.Runtime,
		input.Port,
	).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Slug, &p.RepoFullName, &p.RepoURL,
		&p.Branch, &p.RootDirectory, &p.BuildCommand, &p.StartCommand,
		&p.Runtime, &p.Port, &p.WebhookID, &p.WebhookSecret,
		&p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Store GitHub webhook ID & secret
func SetProjectWebhook(ctx context.Context, id string,
	webhookID int64, secret string) error {
	query := `
		UPDATE projects SET
			webhook_id = $2,
			webhook_secret = $3,
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := pool.Exec(ctx, query, id, webhookID, secret)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("project not found")
	}

	return nil
}

// Remove a project & all related data
func DeleteProject(ctx context.Context, id string) error {
	query := `DELETE FROM projects WHERE id = $1`

	result, err := pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("project not found")
	}

	return nil
}

// Check if a slug is already taken
func SlugExists(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM projects WHERE slug = $1)`

	var exists bool
	err := pool.QueryRow(ctx, query, slug).Scan(&exists)

	return exists, err
}
