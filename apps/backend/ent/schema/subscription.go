// Package schema defines the Ent schema for the application.
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Subscription holds the schema definition for the Subscription entity.
type Subscription struct {
	ent.Schema
}

// Mixin of the Subscription.
func (Subscription) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Fields of the Subscription.
func (Subscription) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("status").
			Values("active", "canceled", "past_due", "trialing").
			Default("active").
			Comment("Subscription status"),
		field.Time("current_period_start").
			Comment("Current billing period start"),
		field.Time("current_period_end").
			Comment("Current billing period end"),
		field.Bool("cancel_at_period_end").
			Default(false).
			Comment("Whether to cancel at period end"),
		// Stripe fields for Phase 14
		field.String("stripe_subscription_id").
			Optional().
			Nillable().
			Comment("Stripe subscription ID"),
		field.String("stripe_customer_id").
			Optional().
			Nillable().
			Comment("Stripe customer ID"),
	}
}

// Edges of the Subscription.
func (Subscription) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("subscriptions").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("plan", Plan.Type).
			Ref("subscriptions").
			Unique().
			Required(),
	}
}

// Indexes of the Subscription.
func (Subscription) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status").
			Edges("user"),
	}
}
