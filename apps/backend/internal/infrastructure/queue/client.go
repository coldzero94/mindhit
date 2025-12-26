// Package queue provides async job queue infrastructure using Asynq.
package queue

import (
	"github.com/hibiken/asynq"
)

// Client wraps asynq.Client for job enqueueing.
type Client struct {
	client *asynq.Client
}

// NewClient creates a new queue client.
func NewClient(redisAddr string) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: redisAddr,
	})
	return &Client{client: client}
}

// Enqueue adds a task to the queue with optional options.
func (c *Client) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.client.Enqueue(task, opts...)
}

// Close closes the client connection.
func (c *Client) Close() error {
	return c.client.Close()
}
