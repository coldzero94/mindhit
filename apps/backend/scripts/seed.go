// Package main provides seed scripts for development database.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/aiconfig"
	"github.com/mindhit/api/ent/mindmapgraph"
	"github.com/mindhit/api/ent/session"

	_ "github.com/lib/pq"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./scripts/seed.go <command>")
		fmt.Println("Commands:")
		fmt.Println("  ai-configs    Create or update AI configs")
		fmt.Println("  mindmaps      Create or update example mindmaps")
		fmt.Println("  all           Run all seeds")
		return fmt.Errorf("no command specified")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/mindhit?sslmode=disable"
	}

	client, err := ent.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	switch os.Args[1] {
	case "ai-configs":
		if err := seedAIConfigs(ctx, client); err != nil {
			return fmt.Errorf("failed to seed ai configs: %w", err)
		}
	case "mindmaps":
		if err := seedMindmaps(ctx, client); err != nil {
			return fmt.Errorf("failed to seed mindmaps: %w", err)
		}
	case "all":
		if err := seedAll(ctx, client); err != nil {
			return fmt.Errorf("failed to seed: %w", err)
		}
	default:
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}

	return nil
}

func seedAll(ctx context.Context, client *ent.Client) error {
	if err := seedAIConfigs(ctx, client); err != nil {
		return err
	}
	if err := seedMindmaps(ctx, client); err != nil {
		return err
	}
	return nil
}

func seedAIConfigs(ctx context.Context, client *ent.Client) error {
	configs := []struct {
		TaskType          string
		Provider          string
		Model             string
		FallbackProviders []string
		Temperature       float64
		MaxTokens         int
		ThinkingBudget    int
		JSONMode          bool
	}{
		{
			TaskType:          "default",
			Provider:          "groq",
			Model:             "llama-3.3-70b-versatile",
			FallbackProviders: []string{"gemini", "openai"},
			Temperature:       0.7,
			MaxTokens:         4096,
			ThinkingBudget:    0,
			JSONMode:          false,
		},
		{
			TaskType:          "tag_extraction",
			Provider:          "groq",
			Model:             "llama-3.1-8b-instant",
			FallbackProviders: []string{"gemini"},
			Temperature:       0.3,
			MaxTokens:         500,
			ThinkingBudget:    0,
			JSONMode:          true,
		},
		{
			TaskType:          "mindmap",
			Provider:          "groq",
			Model:             "llama-3.3-70b-versatile",
			FallbackProviders: []string{"gemini"},
			Temperature:       0.5,
			MaxTokens:         4096,
			ThinkingBudget:    0,
			JSONMode:          true,
		},
	}

	for _, cfg := range configs {
		// Check if config already exists
		exists, err := client.AIConfig.Query().
			Where(aiconfig.TaskTypeEQ(cfg.TaskType)).
			Exist(ctx)
		if err != nil {
			return fmt.Errorf("failed to check ai config %s: %w", cfg.TaskType, err)
		}

		if exists {
			// Update existing config
			_, err = client.AIConfig.Update().
				Where(aiconfig.TaskTypeEQ(cfg.TaskType)).
				SetProvider(cfg.Provider).
				SetModel(cfg.Model).
				SetFallbackProviders(cfg.FallbackProviders).
				SetTemperature(cfg.Temperature).
				SetMaxTokens(cfg.MaxTokens).
				SetThinkingBudget(cfg.ThinkingBudget).
				SetJSONMode(cfg.JSONMode).
				SetEnabled(true).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to update ai config %s: %w", cfg.TaskType, err)
			}
			fmt.Printf("✓ AI config updated: %s\n", cfg.TaskType)
		} else {
			// Create new config
			_, err = client.AIConfig.Create().
				SetTaskType(cfg.TaskType).
				SetProvider(cfg.Provider).
				SetModel(cfg.Model).
				SetFallbackProviders(cfg.FallbackProviders).
				SetTemperature(cfg.Temperature).
				SetMaxTokens(cfg.MaxTokens).
				SetThinkingBudget(cfg.ThinkingBudget).
				SetJSONMode(cfg.JSONMode).
				SetEnabled(true).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create ai config %s: %w", cfg.TaskType, err)
			}
			fmt.Printf("✓ AI config created: %s\n", cfg.TaskType)
		}
	}

	return nil
}

func seedMindmaps(ctx context.Context, client *ent.Client) error {
	// Get or create a session
	sessions, err := client.Session.Query().
		Limit(1).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query sessions: %w", err)
	}

	var sess *ent.Session
	if len(sessions) == 0 {
		// Create a new session
		sess, err = client.Session.Create().
			SetTitle("Sample Browsing Session").
			SetDescription("A sample session with example mindmap data").
			SetSessionStatus(session.SessionStatusCompleted).
			SetStartedAt(time.Now().Add(-time.Hour)).
			SetEndedAt(time.Now()).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
		fmt.Printf("✓ Session created: %s\n", sess.ID)
	} else {
		sess = sessions[0]
		fmt.Printf("✓ Using existing session: %s\n", sess.ID)
	}

	// Check if mindmap already exists for this session
	exists, err := client.MindmapGraph.Query().
		Where(mindmapgraph.HasSessionWith(session.IDEQ(sess.ID))).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("failed to check mindmap: %w", err)
	}

	// Example mindmap data - a galaxy-style mindmap about web development
	nodes := []map[string]interface{}{
		{
			"id":    "core",
			"label": "Web Development",
			"type":  "core",
			"size":  30.0,
			"color": "#3B82F6",
			"position": map[string]interface{}{
				"x": 0.0,
				"y": 0.0,
				"z": 0.0,
			},
			"data": map[string]interface{}{
				"description": "Core topic of the browsing session",
			},
		},
		{
			"id":    "frontend",
			"label": "Frontend",
			"type":  "topic",
			"size":  20.0,
			"color": "#10B981",
			"position": map[string]interface{}{
				"x": 100.0,
				"y": 50.0,
				"z": 30.0,
			},
			"data": map[string]interface{}{},
		},
		{
			"id":    "backend",
			"label": "Backend",
			"type":  "topic",
			"size":  20.0,
			"color": "#F59E0B",
			"position": map[string]interface{}{
				"x": -80.0,
				"y": 70.0,
				"z": -20.0,
			},
			"data": map[string]interface{}{},
		},
		{
			"id":    "react",
			"label": "React",
			"type":  "subtopic",
			"size":  15.0,
			"color": "#61DAFB",
			"position": map[string]interface{}{
				"x": 150.0,
				"y": 100.0,
				"z": 50.0,
			},
			"data": map[string]interface{}{
				"url": "https://react.dev",
			},
		},
		{
			"id":    "nextjs",
			"label": "Next.js",
			"type":  "subtopic",
			"size":  15.0,
			"color": "#000000",
			"position": map[string]interface{}{
				"x": 180.0,
				"y": 30.0,
				"z": 80.0,
			},
			"data": map[string]interface{}{
				"url": "https://nextjs.org",
			},
		},
		{
			"id":    "golang",
			"label": "Go",
			"type":  "subtopic",
			"size":  15.0,
			"color": "#00ADD8",
			"position": map[string]interface{}{
				"x": -120.0,
				"y": 120.0,
				"z": -40.0,
			},
			"data": map[string]interface{}{
				"url": "https://go.dev",
			},
		},
		{
			"id":    "postgres",
			"label": "PostgreSQL",
			"type":  "subtopic",
			"size":  12.0,
			"color": "#336791",
			"position": map[string]interface{}{
				"x": -60.0,
				"y": 150.0,
				"z": -80.0,
			},
			"data": map[string]interface{}{
				"url": "https://postgresql.org",
			},
		},
	}

	edges := []map[string]interface{}{
		{"source": "core", "target": "frontend", "weight": 1.0},
		{"source": "core", "target": "backend", "weight": 1.0},
		{"source": "frontend", "target": "react", "weight": 0.8},
		{"source": "frontend", "target": "nextjs", "weight": 0.8},
		{"source": "backend", "target": "golang", "weight": 0.8},
		{"source": "backend", "target": "postgres", "weight": 0.6},
		{"source": "react", "target": "nextjs", "weight": 0.5, "label": "uses"},
	}

	layout := map[string]interface{}{
		"type": "galaxy",
		"params": map[string]interface{}{
			"spread":  150.0,
			"depth":   100.0,
			"gravity": 0.1,
		},
	}

	if exists {
		// Update existing mindmap
		_, err = client.MindmapGraph.Update().
			Where(mindmapgraph.HasSessionWith(session.IDEQ(sess.ID))).
			SetStatus(mindmapgraph.StatusCompleted).
			SetNodes(nodes).
			SetGraphEdges(edges).
			SetLayout(layout).
			SetGeneratedAt(time.Now()).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update mindmap: %w", err)
		}
		fmt.Printf("✓ Mindmap updated for session: %s\n", sess.ID)
	} else {
		// Create new mindmap
		_, err = client.MindmapGraph.Create().
			SetSessionID(sess.ID).
			SetStatus(mindmapgraph.StatusCompleted).
			SetNodes(nodes).
			SetGraphEdges(edges).
			SetLayout(layout).
			SetGeneratedAt(time.Now()).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create mindmap: %w", err)
		}
		fmt.Printf("✓ Mindmap created for session: %s\n", sess.ID)
	}

	return nil
}
