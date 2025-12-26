package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/internal/service"
	"github.com/mindhit/api/internal/testutil"
)

func setupURLServiceTest(t *testing.T) (*ent.Client, *service.URLService) {
	client := testutil.SetupTestDB(t)
	urlService := service.NewURLService(client)
	return client, urlService
}

// uniqueTestURL generates a unique URL for URL service testing
func uniqueTestURL(prefix string) string {
	return fmt.Sprintf("https://example.com/%s-%s", prefix, uuid.New().String()[:8])
}

// ==================== GetOrCreate Tests ====================

func TestURLService_GetOrCreate_CreateNew(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	rawURL := uniqueTestURL("new")
	title := "Test Page"
	content := "Test content"

	url, err := urlService.GetOrCreate(ctx, rawURL, title, content)

	require.NoError(t, err)
	assert.NotNil(t, url)
	assert.Equal(t, title, url.Title)
	assert.Equal(t, content, url.Content)
	assert.NotEmpty(t, url.URLHash)
}

func TestURLService_GetOrCreate_ReturnExisting(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	rawURL := uniqueTestURL("existing")
	title := "Test Page"
	content := "Test content"

	// Create first
	url1, err := urlService.GetOrCreate(ctx, rawURL, title, content)
	require.NoError(t, err)

	// Get same URL again
	url2, err := urlService.GetOrCreate(ctx, rawURL, "Different Title", "Different content")

	require.NoError(t, err)
	assert.Equal(t, url1.ID, url2.ID)
	assert.Equal(t, title, url2.Title) // Original title preserved
}

func TestURLService_GetOrCreate_UpdateEmptyContent(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	rawURL := uniqueTestURL("update-content")
	title := "Test Page"

	// Create without content
	url1, err := urlService.GetOrCreate(ctx, rawURL, title, "")
	require.NoError(t, err)
	assert.Empty(t, url1.Content)

	// Get again with content - should update
	newContent := "New content"
	url2, err := urlService.GetOrCreate(ctx, rawURL, title, newContent)

	require.NoError(t, err)
	assert.Equal(t, url1.ID, url2.ID)
	assert.Equal(t, newContent, url2.Content)
}

func TestURLService_GetOrCreate_DontOverwriteExistingContent(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	rawURL := uniqueTestURL("keep-content")
	title := "Test Page"
	originalContent := "Original content"

	// Create with content
	url1, err := urlService.GetOrCreate(ctx, rawURL, title, originalContent)
	require.NoError(t, err)

	// Get again with different content - should NOT update
	url2, err := urlService.GetOrCreate(ctx, rawURL, title, "New content")

	require.NoError(t, err)
	assert.Equal(t, url1.ID, url2.ID)
	assert.Equal(t, originalContent, url2.Content) // Original content preserved
}

func TestURLService_GetOrCreate_NormalizeURL(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	base := uuid.New().String()[:8]

	// Create with trailing slash
	url1, err := urlService.GetOrCreate(ctx, fmt.Sprintf("https://example.com/page-%s/", base), "Title", "")
	require.NoError(t, err)

	// Same URL without trailing slash should return same record
	url2, err := urlService.GetOrCreate(ctx, fmt.Sprintf("https://example.com/page-%s", base), "Title", "")

	require.NoError(t, err)
	assert.Equal(t, url1.ID, url2.ID)
}

func TestURLService_GetOrCreate_NormalizeFragment(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	base := uuid.New().String()[:8]

	// Create without fragment
	url1, err := urlService.GetOrCreate(ctx, fmt.Sprintf("https://example.com/page-%s", base), "Title", "")
	require.NoError(t, err)

	// Same URL with fragment should return same record
	url2, err := urlService.GetOrCreate(ctx, fmt.Sprintf("https://example.com/page-%s#section1", base), "Title", "")

	require.NoError(t, err)
	assert.Equal(t, url1.ID, url2.ID)
}

func TestURLService_GetOrCreate_CaseInsensitiveHost(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	base := uuid.New().String()[:8]

	// Create with uppercase host
	url1, err := urlService.GetOrCreate(ctx, fmt.Sprintf("https://EXAMPLE.COM/page-%s", base), "Title", "")
	require.NoError(t, err)

	// Same URL with lowercase host should return same record
	url2, err := urlService.GetOrCreate(ctx, fmt.Sprintf("https://example.com/page-%s", base), "Title", "")

	require.NoError(t, err)
	assert.Equal(t, url1.ID, url2.ID)
}

func TestURLService_GetOrCreate_EmptyPath(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	base := uuid.New().String()[:8]

	// Create with empty path
	url1, err := urlService.GetOrCreate(ctx, fmt.Sprintf("https://%s.example.com", base), "Title", "")
	require.NoError(t, err)

	// Same URL with explicit root path should return same record
	url2, err := urlService.GetOrCreate(ctx, fmt.Sprintf("https://%s.example.com/", base), "Title", "")

	require.NoError(t, err)
	assert.Equal(t, url1.ID, url2.ID)
}

// ==================== GetByHash Tests ====================

func TestURLService_GetByHash_Success(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	rawURL := uniqueTestURL("hash-test")

	// Create URL first
	created, err := urlService.GetOrCreate(ctx, rawURL, "Test", "Content")
	require.NoError(t, err)

	// Get by hash
	found, err := urlService.GetByHash(ctx, created.URLHash)

	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestURLService_GetByHash_NotFound(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	_, err := urlService.GetByHash(ctx, "nonexistent-hash-12345")

	assert.Error(t, err)
}

// ==================== UpdateSummary Tests ====================

func TestURLService_UpdateSummary_Success(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	rawURL := uniqueTestURL("summary")

	// Create URL first
	created, err := urlService.GetOrCreate(ctx, rawURL, "Test", "Content")
	require.NoError(t, err)
	assert.Empty(t, created.Summary)
	assert.Nil(t, created.Keywords)

	// Update summary
	summary := "This is a summary of the page"
	keywords := []string{"test", "example", "page"}

	updated, err := urlService.UpdateSummary(ctx, created.ID, summary, keywords)

	require.NoError(t, err)
	assert.Equal(t, summary, updated.Summary)
	assert.Equal(t, keywords, updated.Keywords)
}

func TestURLService_UpdateSummary_OverwriteExisting(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	rawURL := uniqueTestURL("overwrite-summary")

	// Create URL first
	created, err := urlService.GetOrCreate(ctx, rawURL, "Test", "Content")
	require.NoError(t, err)

	// Set initial summary
	_, err = urlService.UpdateSummary(ctx, created.ID, "First summary", []string{"first"})
	require.NoError(t, err)

	// Update with new summary
	newSummary := "Second summary"
	newKeywords := []string{"second", "updated"}

	updated, err := urlService.UpdateSummary(ctx, created.ID, newSummary, newKeywords)

	require.NoError(t, err)
	assert.Equal(t, newSummary, updated.Summary)
	assert.Equal(t, newKeywords, updated.Keywords)
}

func TestURLService_UpdateSummary_NotFound(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	nonExistentID := uuid.New()

	_, err := urlService.UpdateSummary(ctx, nonExistentID, "Summary", []string{"keyword"})

	assert.Error(t, err)
}

// ==================== GetURLsWithoutSummary Tests ====================

func TestURLService_GetURLsWithoutSummary_Success(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	// Create URL with content but no summary
	url1, err := urlService.GetOrCreate(ctx, uniqueTestURL("no-summary-1"), "Test 1", "Content 1")
	require.NoError(t, err)
	_ = url1

	// Create URL with content and summary
	url2, err := urlService.GetOrCreate(ctx, uniqueTestURL("with-summary"), "Test 2", "Content 2")
	require.NoError(t, err)
	_, err = urlService.UpdateSummary(ctx, url2.ID, "Summary", []string{"keyword"})
	require.NoError(t, err)

	// Create URL with content but no summary
	url3, err := urlService.GetOrCreate(ctx, uniqueTestURL("no-summary-2"), "Test 3", "Content 3")
	require.NoError(t, err)
	_ = url3

	// Get URLs without summary
	urls, err := urlService.GetURLsWithoutSummary(ctx, 10)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(urls), 2) // At least url1 and url3
}

func TestURLService_GetURLsWithoutSummary_LimitWorks(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	// Create 5 URLs with content but no summary
	for i := 0; i < 5; i++ {
		_, err := urlService.GetOrCreate(ctx, uniqueTestURL(fmt.Sprintf("limit-test-%d", i)), "Test", "Content")
		require.NoError(t, err)
	}

	// Get with limit of 2
	urls, err := urlService.GetURLsWithoutSummary(ctx, 2)

	require.NoError(t, err)
	assert.Len(t, urls, 2)
}

func TestURLService_GetURLsWithoutSummary_ExcludesEmptyContent(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	// Create URL without content (should not be returned)
	urlWithoutContent, err := urlService.GetOrCreate(ctx, uniqueTestURL("empty-content"), "Test", "")
	require.NoError(t, err)

	// Create URL with content but no summary (should be returned)
	urlWithContent, err := urlService.GetOrCreate(ctx, uniqueTestURL("has-content"), "Test", "Content")
	require.NoError(t, err)

	// Get URLs without summary (use high limit to get all)
	urls, err := urlService.GetURLsWithoutSummary(ctx, 100)

	require.NoError(t, err)

	// Check that URL with content is included
	foundWithContent := false
	foundWithoutContent := false
	for _, u := range urls {
		if u.ID == urlWithContent.ID {
			foundWithContent = true
		}
		if u.ID == urlWithoutContent.ID {
			foundWithoutContent = true
		}
	}
	assert.True(t, foundWithContent, "Expected URL with content to be in results")
	assert.False(t, foundWithoutContent, "URL without content should not be in results")
}

func TestURLService_GetURLsWithoutSummary_Empty(t *testing.T) {
	client, urlService := setupURLServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	// Create URL with summary only
	url, err := urlService.GetOrCreate(ctx, uniqueTestURL("summarized"), "Test", "Content")
	require.NoError(t, err)
	_, err = urlService.UpdateSummary(ctx, url.ID, "Summary", []string{"keyword"})
	require.NoError(t, err)

	// Get URLs without summary - may have other URLs from shared test DB
	urls, err := urlService.GetURLsWithoutSummary(ctx, 10)

	require.NoError(t, err)

	// The summarized URL should not be in the results
	for _, u := range urls {
		assert.NotEqual(t, url.ID, u.ID, "Summarized URL should not appear in results")
	}
}
