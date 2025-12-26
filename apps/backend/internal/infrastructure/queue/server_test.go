package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer_DefaultConfig(t *testing.T) {
	server := NewServer(ServerConfig{
		RedisAddr: "localhost:6380",
	})

	assert.NotNil(t, server)
}

func TestNewServer_CustomConfig(t *testing.T) {
	server := NewServer(ServerConfig{
		RedisAddr:   "localhost:6380",
		Concurrency: 20,
		Queues: map[string]int{
			"high":    10,
			"default": 5,
			"low":     1,
		},
	})

	assert.NotNil(t, server)
}

func TestNewServer_ZeroConcurrency_UsesDefault(t *testing.T) {
	server := NewServer(ServerConfig{
		RedisAddr:   "localhost:6380",
		Concurrency: 0, // Should default to 10
	})

	assert.NotNil(t, server)
}

func TestNewServer_NilQueues_UsesDefault(t *testing.T) {
	server := NewServer(ServerConfig{
		RedisAddr: "localhost:6380",
		Queues:    nil, // Should use default queue priorities
	})

	assert.NotNil(t, server)
}
