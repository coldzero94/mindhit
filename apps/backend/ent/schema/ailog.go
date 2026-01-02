// Package schema defines the Ent schema for the application.
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AILog holds the schema definition for the AILog entity.
type AILog struct {
	ent.Schema
}

// Fields of the AILog.
func (AILog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),

		// Relations (optional)
		field.UUID("user_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("session_id", uuid.UUID{}).
			Optional().
			Nillable(),

		// Request info
		field.String("task_type").
			NotEmpty().
			Comment("tag_extraction, mindmap, general"),
		field.String("provider").
			NotEmpty().
			Comment("openai, claude, gemini"),
		field.String("model").
			NotEmpty(),
		field.Text("system_prompt").
			Optional(),
		field.Text("user_prompt").
			Optional(),

		// Response info
		field.Text("thinking").
			Optional().
			Comment("AI reasoning/thinking process"),
		field.Text("content").
			Optional().
			Comment("AI response content (empty on error)"),

		// Token usage (accurate values from API response)
		field.Int("input_tokens").
			Default(0),
		field.Int("output_tokens").
			Default(0),
		field.Int("thinking_tokens").
			Default(0),
		field.Int("total_tokens").
			Default(0),

		// Performance metrics
		field.Int64("latency_ms").
			Default(0).
			Comment("Response latency in milliseconds"),
		field.String("request_id").
			Optional().
			Comment("Provider request ID for debugging"),

		// Status
		field.Enum("status").
			Values("success", "error", "timeout").
			Default("success"),
		field.Text("error_message").
			Optional(),

		// Cost tracking (cents)
		field.Int("estimated_cost_cents").
			Default(0).
			Comment("Estimated cost in cents"),

		// Additional metadata
		field.JSON("metadata", map[string]interface{}{}).
			Optional().
			Comment("Additional tracking metadata"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Indexes of the AILog.
func (AILog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "created_at"),
		index.Fields("session_id"),
		index.Fields("task_type", "created_at"),
		index.Fields("provider", "model", "created_at"),
		index.Fields("status", "created_at"),
	}
}

// Edges of the AILog.
func (AILog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("ai_logs").
			Field("user_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("session", Session.Type).
			Ref("ai_logs").
			Field("session_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
