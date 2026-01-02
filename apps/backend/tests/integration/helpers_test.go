//go:build integration

package integration

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

var emailCounter uint64

// uniqueEmail generates a unique email for testing to avoid conflicts.
func uniqueEmail(prefix string) string {
	count := atomic.AddUint64(&emailCounter, 1)
	return fmt.Sprintf("%s_%d_%d@test.local", prefix, time.Now().UnixNano(), count)
}

// parseUUID parses a string to UUID.
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
