// Package schema defines the Ent schema for the application.
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AIConfig holds dynamic AI provider configuration.
// API keys are managed via environment variables, only provider/model selection in DB.
type AIConfig struct {
	ent.Schema
}

// Fields of the AIConfig.
func (AIConfig) Fields() []ent.Field {
	return []ent.Field{
		// Task type (unique key)
		field.String("task_type").
			NotEmpty().
			Unique().
			Comment("Task type: 'default', 'tag_extraction', 'mindmap'"),

		// Provider settings
		field.String("provider").
			NotEmpty().
			Comment("AI provider: 'openai', 'claude', 'gemini'"),
		field.String("model").
			NotEmpty().
			Comment("Model name: 'gpt-4o', 'claude-sonnet-4', 'gemini-2.0-flash'"),

		// Fallback providers
		field.JSON("fallback_providers", []string{}).
			Optional().
			Comment("Ordered list of fallback providers"),

		// Options
		field.Float("temperature").
			Default(0.7).
			Comment("Model temperature (0.0-2.0)"),
		field.Int("max_tokens").
			Default(4096).
			Comment("Max output tokens"),
		field.Int("thinking_budget").
			Default(0).
			Comment("Extended thinking token budget (Claude)"),
		field.Bool("json_mode").
			Default(false).
			Comment("Force JSON output"),

		// Enable status
		field.Bool("enabled").
			Default(true).
			Comment("Whether this config is active"),

		// Audit fields
		field.String("updated_by").
			Optional().
			Comment("Admin who last updated this config"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Indexes of the AIConfig.
func (AIConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("task_type").Unique(),
		index.Fields("provider", "model"),
	}
}
