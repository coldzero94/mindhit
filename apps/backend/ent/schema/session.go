package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Session struct {
	ent.Schema
}

func (Session) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
		SoftDeleteMixin{},
	}
}

func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").
			Optional().
			Nillable().
			Comment("Session title"),
		field.Text("description").
			Optional().
			Nillable().
			Comment("Session description"),
		field.Enum("session_status").
			Values("recording", "paused", "processing", "completed", "failed").
			Default("recording").
			Comment("Session workflow status"),
		field.Time("started_at").
			Default(time.Now).
			Comment("Session start time"),
		field.Time("ended_at").
			Optional().
			Nillable().
			Comment("Session end time"),
	}
}

func (Session) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("sessions").
			Unique().
			Required(),
		edge.To("page_visits", PageVisit.Type),
		edge.To("highlights", Highlight.Type),
		edge.To("raw_events", RawEvent.Type),
		edge.To("mindmap", MindmapGraph.Type).
			Unique(),
		edge.To("token_usage", TokenUsage.Type),
	}
}

func (Session) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_status"),
	}
}
