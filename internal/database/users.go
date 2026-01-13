package database

import (
	"context"
	"errors"
	"time"

	"github.com/Sys-Redux/rcnbuild-paas/pkg/crypto"
)

// User represents a user in the database
type User struct {
	ID                   string    `json:"id"`
	GitHubID             int64     `json:"github_id"`
	GitHubUsername       string    `json:"github_username"`
	Email                *string   `json:"email,omitempty"`
	AvatarURL            *string   `json:"avatar_url,omitempty"`
	AccessTokenEncrypted *string   `json:"-"` // Never expose in JSON
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// GitHubUser represents the user info returned from GitHub API
type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// Upserts a user based on GitHub ID
func CreateOrUpdateUser(
	ctx context.Context, githubUser *GitHubUser,
	accessToken string) (*User, error) {
	query := `
	INSERT INTO users (
		github_id,
		github_username,
		email,
		avatar_url,
		access_token_encrypted,
		updated_at)
	VALUES ($1, $2, $3, $4, $5, NOW())
	ON CONFLICT (github_id) DO UPDATE SET
		github_username = EXCLUDED.github_username,
		email = EXCLUDED.email,
		avatar_url = EXCLUDED.avatar_url,
		access_token_encrypted = EXCLUDED.access_token_encrypted,
		updated_at = NOW()
	RETURNING id, github_id, github_username, email, avatar_url, created_at, updated_at
	`

	// Encrypt access token before storing
	encryptedToken, err := crypto.Encrypt(accessToken)
	if err != nil {
		return nil, err
	}

	var user User
	var email, avatarURL *string

	if githubUser.Email != "" {
		email = &githubUser.Email
	}
	if githubUser.AvatarURL != "" {
		avatarURL = &githubUser.AvatarURL
	}

	err = pool.QueryRow(ctx, query,
		githubUser.ID,
		githubUser.Login,
		email,
		avatarURL,
		encryptedToken,
	).Scan(
		&user.ID,
		&user.GitHubID,
		&user.GitHubUsername,
		&user.Email,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Retrieves user by their UUID
func GetUserByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT
			id,
			github_id,
			github_username,
			email,
			avatar_url,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.GitHubID,
		&user.GitHubUsername,
		&user.Email,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Retrieves a user by their GitHub ID
func GetUserByGitHubID(ctx context.Context, githubID int64) (*User, error) {
	query := `
		SELECT
			id,
			github_id,
			github_username,
			email,
			avatar_url,
			created_at,
			updated_at
		FROM users
		WHERE github_id = $1
	`

	var user User
	err := pool.QueryRow(ctx, query, githubID).Scan(
		&user.ID,
		&user.GitHubID,
		&user.GitHubUsername,
		&user.Email,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// DeleteUser permanently removes a user by their UUID
func DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}

// UpdateUserEmail updates just the user's email
func UpdateUserEmail(ctx context.Context, id string, email string) error {
	query := `
		UPDATE users
		SET email = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := pool.Exec(ctx, query, id, email)
	return err
}

// GetUserAccessToken retrieves and decrypts the GitHub access token for a user
// Used internally for making GitHub API calls on behalf of the user
func GetUserAccessToken(ctx context.Context, userID string) (string, error) {
	query := `
		SELECT access_token_encrypted
		FROM users
		WHERE id = $1
	`

	var encryptedToken *string
	err := pool.QueryRow(ctx, query, userID).Scan(&encryptedToken)
	if err != nil {
		return "", err
	}

	if encryptedToken == nil {
		return "", errors.New("user has no access token")
	}

	return crypto.Decrypt(*encryptedToken)
}
