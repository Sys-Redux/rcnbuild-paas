package queue

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

// Asynq client for enqueueing jobs
var client *asynq.Client

// Initialize asynq client
func Connect(redisAddr string) error {
	client = asynq.NewClient(asynq.RedisClientOpt{
		Addr: redisAddr,
	})
	log.Info().Str("redis_addr", redisAddr).Msg("Connected to Asynq client")
	return nil
}

// Close asynq client
func Close() error {
	if client != nil {
		return client.Close()
	}
	return nil
}

// Enqueue a job
func EnqueueBuild(ctx context.Context, payload *BuildPayload) (string, error) {
	task, err := NewBuildTask(payload)
	if err != nil {
		return "", err
	}

	info, err := client.EnqueueContext(ctx, task)
	if err != nil {
		return "", err
	}

	log.Info().
		Str("task_id", info.ID).
		Str("queue", info.Queue).
		Str("deployment_id", payload.DeploymentID).
		Msg("Enqueued build job")

	return info.ID, nil
}

// Enqueue a Deploy job
func EnqueueDeploy(ctx context.Context,
	payload *DeployPayload) (string, error) {
	task, err := NewDeployTask(payload)
	if err != nil {
		return "", err
	}

	info, err := client.EnqueueContext(ctx, task)
	if err != nil {
		return "", err
	}

	log.Info().
		Str("task_id", info.ID).
		Str("queue", info.Queue).
		Str("deployment_id", payload.DeploymentID).
		Msg("Enqueued deploy job")

	return info.ID, nil
}
