package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		SoftDeleteMixin{},
	}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").
			Unique().
			NotEmpty().
			Comment("User email address"),
		field.String("password_hash").
			Sensitive().
			Comment("Hashed password"),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("settings", UserSettings.Type).
			Unique(),
		edge.To("sessions", Session.Type),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
		index.Fields("status"),
	}
}
