package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Mixin returns the mixins for User.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields for User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").
			Unique().
			NotEmpty().
			Comment("User email address"),
		field.String("password_hash").
			Optional().
			Nillable().
			Sensitive().
			Comment("Hashed password - nil for Google OAuth users"),

		// OAuth fields
		field.String("google_id").
			Optional().
			Unique().
			Nillable().
			Comment("Google user ID from OAuth"),
		field.String("avatar_url").
			Optional().
			Nillable().
			Comment("Profile picture URL"),
		field.Enum("auth_provider").
			Values("email", "google").
			Default("email").
			Comment("Authentication provider used for signup"),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("settings", UserSettings.Type).
			Unique(),
		edge.To("sessions", Session.Type),
		edge.To("password_reset_tokens", PasswordResetToken.Type),
		edge.To("subscriptions", Subscription.Type),
		edge.To("token_usage", TokenUsage.Type),
		edge.To("ai_logs", AILog.Type),
	}
}

// Indexes returns the indexes for User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
		index.Fields("status"),
		index.Fields("google_id"),
		index.Fields("auth_provider"),
	}
}
