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

// PasswordResetToken holds the schema definition for password reset tokens.
type PasswordResetToken struct {
	ent.Schema
}

func (PasswordResetToken) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("token").
			Unique().
			NotEmpty().
			Comment("Secure reset token"),
		field.UUID("user_id", uuid.UUID{}).
			Comment("Owner user ID"),
		field.Time("expires_at").
			Comment("Token expiration time"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("Token creation time"),
		field.Bool("used").
			Default(false).
			Comment("Whether the token has been used"),
	}
}

func (PasswordResetToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("password_reset_tokens").
			Field("user_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (PasswordResetToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token"),
		index.Fields("user_id"),
		index.Fields("expires_at"),
	}
}
