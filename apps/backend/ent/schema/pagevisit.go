package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type PageVisit struct {
	ent.Schema
}

func (PageVisit) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (PageVisit) Fields() []ent.Field {
	return []ent.Field{
		field.Time("entered_at").
			Default(time.Now).
			Comment("Page entry time"),
		field.Time("left_at").
			Optional().
			Nillable().
			Comment("Page leave time"),
		field.Int("duration_ms").
			Optional().
			Nillable().
			Comment("Time spent on page in milliseconds"),
		field.Float("max_scroll_depth").
			Default(0).
			Comment("Maximum scroll depth (0-1)"),
	}
}

func (PageVisit) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("session", Session.Type).
			Ref("page_visits").
			Unique().
			Required(),
		edge.To("url", URL.Type).
			Unique().
			Required(),
	}
}

func (PageVisit) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("entered_at"),
	}
}
