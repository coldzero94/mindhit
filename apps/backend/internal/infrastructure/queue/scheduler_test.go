package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewScheduler(t *testing.T) {
	scheduler, err := NewScheduler("localhost:6380")

	require.NoError(t, err)
	assert.NotNil(t, scheduler)
}

func TestNewScheduler_DifferentAddresses(t *testing.T) {
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
			scheduler, err := NewScheduler(tt.addr)
			require.NoError(t, err)
			assert.NotNil(t, scheduler)
		})
	}
}
