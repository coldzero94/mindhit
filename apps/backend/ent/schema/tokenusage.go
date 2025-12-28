// Package schema defines the Ent schema for the application.
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// TokenUsage holds the schema definition for the TokenUsage entity.
type TokenUsage struct {
	ent.Schema
}

// Mixin of the TokenUsage.
func (TokenUsage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Fields of the TokenUsage.
func (TokenUsage) Fields() []ent.Field {
	return []ent.Field{
		field.String("operation").
			NotEmpty().
			Comment("AI operation type: 'summarize', 'mindmap', 'keywords'"),
		field.Int("tokens_used").
			Positive().
			Comment("Number of tokens used"),
		field.String("ai_model").
			Optional().
			Comment("AI model used (e.g., 'gpt-4', 'claude-3')"),
		field.Time("period_start").
			Comment("Billing period this usage belongs to"),
	}
}

// Edges of the TokenUsage.
func (TokenUsage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("token_usage").
			Unique().
			Required(),
		edge.From("session", Session.Type).
			Ref("token_usage").
			Unique(),
	}
}

// Indexes of the TokenUsage.
func (TokenUsage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("period_start").
			Edges("user"),
		index.Fields("operation").
			Edges("user"),
	}
}
