package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type RawEvent struct {
	ent.Schema
}

func (RawEvent) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (RawEvent) Fields() []ent.Field {
	return []ent.Field{
		field.String("event_type").
			NotEmpty().
			Comment("Event type (page_visit, highlight, scroll, etc.)"),
		field.Time("timestamp").
			Comment("Client-side event timestamp"),
		field.Text("payload").
			Comment("Raw JSON event payload"),
		field.Bool("processed").
			Default(false).
			Comment("Whether event has been processed"),
		field.Time("processed_at").
			Optional().
			Nillable().
			Comment("When event was processed"),
	}
}

func (RawEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("session", Session.Type).
			Ref("raw_events").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (RawEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_type"),
		index.Fields("processed"),
		index.Fields("timestamp"),
	}
}
