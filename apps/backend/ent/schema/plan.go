// Package schema defines the Ent schema for the application.
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Plan holds the schema definition for the Plan entity.
type Plan struct {
	ent.Schema
}

// Fields of the Plan.
func (Plan) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable().
			Comment("Plan identifier: 'free', 'pro', 'enterprise'"),
		field.String("name").
			NotEmpty().
			Comment("Display name of the plan"),
		field.Int("price_cents").
			Default(0).
			Comment("Monthly price in cents (0 for free, null for custom)"),
		field.String("billing_period").
			Default("monthly").
			Comment("Billing period: 'monthly' or 'yearly'"),
		field.Int("token_limit").
			Optional().
			Nillable().
			Comment("Monthly token limit (null = unlimited)"),
		field.Int("session_retention_days").
			Optional().
			Nillable().
			Comment("Days to retain sessions (null = unlimited)"),
		field.Int("max_concurrent_sessions").
			Optional().
			Nillable().
			Comment("Max concurrent sessions (null = unlimited)"),
		field.JSON("features", map[string]bool{}).
			Default(map[string]bool{}).
			Comment("Feature flags for this plan"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("Plan creation time"),
	}
}

// Edges of the Plan.
func (Plan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("subscriptions", Subscription.Type),
	}
}
