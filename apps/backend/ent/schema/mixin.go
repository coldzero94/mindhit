package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// BaseMixin defines common fields for all entities
type BaseMixin struct {
	mixin.Schema
}

func (BaseMixin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable().
			Comment("Primary key"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("Record creation timestamp"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Record last update timestamp"),
	}
}

// SoftDeleteMixin adds soft delete capability
type SoftDeleteMixin struct {
	mixin.Schema
}

func (SoftDeleteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("status").
			Values("active", "inactive").
			Default("active").
			Comment("Record status for soft delete"),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Soft delete timestamp"),
	}
}
