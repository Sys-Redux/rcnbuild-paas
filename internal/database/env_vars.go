package database

import (
	"context"
	"errors"
	"time"
)

// Represents environment variables for a project (internal use only)
// NEVER return this struct directly in API responses - use ToDisplay()
//
// Security model:
// - Database stores ONLY encrypted values (value_encrypted column)
// - Decryption happens ONLY in GetEnvVarsAsMap() for Docker container injection
// - API responses use ToDisplay() which masks all values
type EnvVar struct {
	ID             string    `json:"-"` // Prevent accidental serialization
	ProjectID      string    `json:"-"`
	Key            string    `json:"-"`
	ValueEncrypted string    `json:"-"`
	CreatedAt      time.Time `json:"-"`
}

// Safe for API responses - values are always masked per API spec
type EnvVarDisplay struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"` // Always masked: "••••••••"
	CreatedAt time.Time `json:"created_at"`
}

const maskedValue = "••••••••"

// ToDisplay converts to a safe API response format with masked value
func (e *EnvVar) ToDisplay() EnvVarDisplay {
	return EnvVarDisplay{
		ID:        e.ID,
		Key:       e.Key,
		Value:     maskedValue,
		CreatedAt: e.CreatedAt,
	}
}

// ToDisplayList converts a slice of EnvVars to safe display format
func ToDisplayList(envVars []*EnvVar) []EnvVarDisplay {
	result := make([]EnvVarDisplay, len(envVars))
	for i, e := range envVars {
		result[i] = e.ToDisplay()
	}
	return result
}

// Upserts an environment variable for a project
// NOTE: Caller must encrypt value first using crypto.Encrypt()
func CreateOrUpdateEnvVar(ctx context.Context, projectID, key,
	encryptedValue string) (*EnvVar, error) {
	query := `
		INSERT INTO env_vars (
			project_id, key, value_encrypted
		) VALUES ($1, $2, $3)
		ON CONFLICT (project_id, key) DO UPDATE SET
			value_encrypted = EXCLUDED.value_encrypted
		RETURNING id, project_id, key, value_encrypted, created_at
	`

	var e EnvVar
	err := pool.QueryRow(ctx, query,
		projectID, key, encryptedValue,
	).Scan(
		&e.ID, &e.ProjectID, &e.Key, &e.ValueEncrypted, &e.CreatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &e, nil
}

// Returns all environment variables for a project
func GetEnvVarsByProjectID(ctx context.Context,
	projectID string) ([]*EnvVar, error) {
	query := `
		SELECT id, project_id, key, value_encrypted, created_at
		FROM env_vars
		WHERE project_id = $1
		ORDER BY created_at ASC
	`

	rows, err := pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envVars []*EnvVar
	for rows.Next() {
		var e EnvVar
		err := rows.Scan(
			&e.ID, &e.ProjectID, &e.Key, &e.ValueEncrypted, &e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		envVars = append(envVars, &e)
	}

	return envVars, nil
}

// Removes an environment variable
func DeleteEnvVar(ctx context.Context, projectID, key string) error {
	query := `
		DELETE FROM env_vars
		WHERE project_id = $1 AND key = $2
	`

	result, err := pool.Exec(ctx, query, projectID, key)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("env var not found")
	}

	return nil
}

// Removes all environment variables for a project
func DeleteAllEnvVar(ctx context.Context, projectID string) error {
	query := `
		DELETE FROM env_vars
		WHERE project_id = $1
	`

	_, err := pool.Exec(ctx, query, projectID)
	return err
}

// Returns environment variables as a map for container injection
// Values are decrypted before returning
// Usage: database.GetEnvVarsAsMap(ctx, projectID, crypto.Decrypt)
func GetEnvVarsAsMap(ctx context.Context, projectID string,
	decryptFn func(string) (string, error)) (map[string]string, error) {
	envVars, err := GetEnvVarsByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, e := range envVars {
		decrypted, err := decryptFn(e.ValueEncrypted)
		if err != nil {
			return nil, err
		}
		result[e.Key] = decrypted
	}

	return result, nil
}
