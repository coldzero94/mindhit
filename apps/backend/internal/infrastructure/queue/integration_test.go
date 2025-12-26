package queue

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestRedisAddr returns the Redis address for testing
func getTestRedisAddr() string {
	if addr := os.Getenv("TEST_REDIS_ADDR"); addr != "" {
		return addr
	}
	return "localhost:6380"
}

// skipIfNoRedis skips the test if Redis is not available
func skipIfNoRedis(t *testing.T) {
	t.Helper()
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: getTestRedisAddr()})
	defer func() { _ = client.Close() }()

	// Try to ping Redis by enqueueing and immediately deleting a test task
	task := asynq.NewTask("test:ping", nil)
	info, err := client.Enqueue(task)
	if err != nil {
		t.Skipf("Skipping test: Redis not available at %s: %v", getTestRedisAddr(), err)
	}
	// Clean up the test task
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: getTestRedisAddr()})
	defer func() { _ = inspector.Close() }()
	_ = inspector.DeleteTask("default", info.ID)
}

func TestClient_Enqueue_Integration(t *testing.T) {
	skipIfNoRedis(t)

	client := NewClient(getTestRedisAddr())
	defer func() { _ = client.Close() }()

	task, err := NewSessionProcessTask("test-session-123")
	require.NoError(t, err)

	info, err := client.Enqueue(task, asynq.MaxRetry(3))

	require.NoError(t, err)
	assert.NotEmpty(t, info.ID)
	assert.Equal(t, TypeSessionProcess, info.Type)

	// Clean up
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: getTestRedisAddr()})
	defer func() { _ = inspector.Close() }()
	_ = inspector.DeleteTask(info.Queue, info.ID)
}

func TestClient_EnqueueMultipleTasks_Integration(t *testing.T) {
	skipIfNoRedis(t)

	client := NewClient(getTestRedisAddr())
	defer func() { _ = client.Close() }()

	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: getTestRedisAddr()})
	defer func() { _ = inspector.Close() }()

	var taskIDs []string

	// Enqueue multiple tasks
	for i := 0; i < 3; i++ {
		task, err := NewSessionProcessTask("session-" + string(rune('A'+i)))
		require.NoError(t, err)

		info, err := client.Enqueue(task)
		require.NoError(t, err)
		taskIDs = append(taskIDs, info.ID)
	}

	assert.Len(t, taskIDs, 3)

	// Clean up
	for _, id := range taskIDs {
		_ = inspector.DeleteTask("default", id)
	}
}

func TestScheduler_RegisterPeriodicTasks_Integration(t *testing.T) {
	skipIfNoRedis(t)

	scheduler, err := NewScheduler(getTestRedisAddr())
	require.NoError(t, err)

	err = scheduler.RegisterPeriodicTasks()
	require.NoError(t, err)

	// Scheduler should be able to shut down cleanly
	scheduler.Shutdown()
}

func TestServer_Creation_Integration(t *testing.T) {
	skipIfNoRedis(t)

	server := NewServer(ServerConfig{
		RedisAddr:   getTestRedisAddr(),
		Concurrency: 5,
	})

	// Register a test handler
	handled := make(chan bool, 1)
	server.HandleFunc("test:task", func(_ context.Context, _ *asynq.Task) error {
		handled <- true
		return nil
	})

	// Server should be created successfully
	assert.NotNil(t, server)

	// Clean shutdown
	server.Shutdown()
}

func TestEndToEnd_EnqueueAndProcess_Integration(t *testing.T) {
	skipIfNoRedis(t)

	redisAddr := getTestRedisAddr()

	// Create client
	client := NewClient(redisAddr)
	defer func() { _ = client.Close() }()

	// Create server with test handler
	server := NewServer(ServerConfig{
		RedisAddr:   redisAddr,
		Concurrency: 1,
	})

	processed := make(chan string, 1)
	server.HandleFunc("test:e2e", func(_ context.Context, task *asynq.Task) error {
		processed <- string(task.Payload())
		return nil
	})

	// Start server in background
	go func() {
		_ = server.Run()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Enqueue task
	task := asynq.NewTask("test:e2e", []byte("hello-world"))
	_, err := client.Enqueue(task)
	require.NoError(t, err)

	// Wait for processing
	select {
	case payload := <-processed:
		assert.Equal(t, "hello-world", payload)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for task to be processed")
	}

	// Shutdown
	server.Shutdown()
}

func TestClient_EnqueueWithOptions_Integration(t *testing.T) {
	skipIfNoRedis(t)

	client := NewClient(getTestRedisAddr())
	defer func() { _ = client.Close() }()

	task, err := NewSessionCleanupTask(24)
	require.NoError(t, err)

	// Enqueue with various options
	info, err := client.Enqueue(task,
		asynq.MaxRetry(5),
		asynq.Queue("critical"),
		asynq.Timeout(30*time.Second),
	)

	require.NoError(t, err)
	assert.Equal(t, "critical", info.Queue)
	assert.Equal(t, 5, info.MaxRetry)

	// Clean up
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: getTestRedisAddr()})
	defer func() { _ = inspector.Close() }()
	_ = inspector.DeleteTask(info.Queue, info.ID)
}

func TestClient_EnqueueScheduled_Integration(t *testing.T) {
	skipIfNoRedis(t)

	client := NewClient(getTestRedisAddr())
	defer func() { _ = client.Close() }()

	task, err := NewMindmapGenerateTask("session-scheduled")
	require.NoError(t, err)

	// Schedule for 1 hour from now
	info, err := client.Enqueue(task, asynq.ProcessIn(time.Hour))

	require.NoError(t, err)
	assert.NotEmpty(t, info.ID)

	// Clean up
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: getTestRedisAddr()})
	defer func() { _ = inspector.Close() }()
	_ = inspector.DeleteTask(info.Queue, info.ID)
}
