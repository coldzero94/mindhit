package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type UserSettings struct {
	ent.Schema
}

func (UserSettings) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (UserSettings) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("theme").
			Values("light", "dark", "system").
			Default("system").
			Comment("UI theme preference"),
		field.Bool("email_notifications").
			Default(true).
			Comment("Email notification preference"),
		field.Bool("browser_notifications").
			Default(true).
			Comment("Browser notification preference"),
		field.String("language").
			Default("ko").
			Comment("Preferred language"),
		field.Int("session_timeout_minutes").
			Default(60).
			Comment("Auto-stop session after inactivity"),
		field.Bool("auto_summarize").
			Default(true).
			Comment("Auto-generate summary when session ends"),
		field.JSON("extension_settings", map[string]interface{}{}).
			Optional().
			Comment("Chrome extension specific settings"),
	}
}

func (UserSettings) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("settings").
			Unique().
			Required(),
	}
}
