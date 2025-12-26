// Package service provides business logic implementations.
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/mindhit/api/ent"
	enturl "github.com/mindhit/api/ent/url"
)

// URLService handles URL-related business logic.
type URLService struct {
	client *ent.Client
}

// NewURLService creates a new URLService instance.
func NewURLService(client *ent.Client) *URLService {
	return &URLService{client: client}
}

// GetOrCreate retrieves an existing URL or creates a new one.
func (s *URLService) GetOrCreate(
	ctx context.Context,
	rawURL string,
	title string,
	content string,
) (*ent.URL, error) {
	// Normalize URL
	normalizedURL := normalizeURL(rawURL)
	urlHash := hashURL(normalizedURL)

	// Try to find existing URL
	existing, err := s.client.URL.
		Query().
		Where(enturl.URLHashEQ(urlHash)).
		Only(ctx)

	if err == nil {
		// Update content if provided and not already set
		if content != "" && existing.Content == "" {
			return s.client.URL.
				UpdateOne(existing).
				SetContent(content).
				Save(ctx)
		}
		return existing, nil
	}

	if !ent.IsNotFound(err) {
		return nil, err
	}

	// Create new URL
	create := s.client.URL.
		Create().
		SetURL(normalizedURL).
		SetURLHash(urlHash)

	if title != "" {
		create.SetTitle(title)
	}
	if content != "" {
		create.SetContent(content)
	}

	return create.Save(ctx)
}

// GetByHash retrieves a URL by its hash.
func (s *URLService) GetByHash(ctx context.Context, urlHash string) (*ent.URL, error) {
	return s.client.URL.
		Query().
		Where(enturl.URLHashEQ(urlHash)).
		Only(ctx)
}

// UpdateSummary updates the AI-generated summary for a URL.
func (s *URLService) UpdateSummary(
	ctx context.Context,
	urlID uuid.UUID,
	summary string,
	keywords []string,
) (*ent.URL, error) {
	return s.client.URL.
		UpdateOneID(urlID).
		SetSummary(summary).
		SetKeywords(keywords).
		Save(ctx)
}

// GetURLsWithoutSummary retrieves URLs that need summarization.
func (s *URLService) GetURLsWithoutSummary(ctx context.Context, limit int) ([]*ent.URL, error) {
	return s.client.URL.
		Query().
		Where(
			enturl.ContentNotNil(),
			enturl.SummaryIsNil(),
		).
		Limit(limit).
		All(ctx)
}

// normalizeURL removes fragments and normalizes the URL.
func normalizeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Remove fragment
	parsed.Fragment = ""

	// Normalize path
	if parsed.Path == "" {
		parsed.Path = "/"
	}

	// Remove trailing slash for non-root paths
	if parsed.Path != "/" && strings.HasSuffix(parsed.Path, "/") {
		parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	}

	// Lowercase scheme and host
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)

	return parsed.String()
}

// hashURL creates a SHA256 hash of the URL.
func hashURL(normalizedURL string) string {
	hash := sha256.Sum256([]byte(normalizedURL))
	return hex.EncodeToString(hash[:])
}
