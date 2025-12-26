package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	// Test that client can be created (doesn't require actual Redis connection)
	client := NewClient("localhost:6380")
	assert.NotNil(t, client)

	err := client.Close()
	assert.NoError(t, err)
}

func TestNewClient_DifferentAddresses(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{"localhost", "localhost:6379"},
		{"custom port", "localhost:6380"},
		{"remote host", "redis.example.com:6379"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.addr)
			assert.NotNil(t, client)
			_ = client.Close()
		})
	}
}
