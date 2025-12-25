package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type MindmapGraph struct {
	ent.Schema
}

func (MindmapGraph) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (MindmapGraph) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("nodes", []map[string]interface{}{}).
			Optional().
			Comment("Mindmap node data"),
		field.JSON("graph_edges", []map[string]interface{}{}).
			Optional().
			Comment("Mindmap edge data"),
		field.JSON("layout", map[string]interface{}{}).
			Optional().
			Comment("Layout configuration"),
		field.Time("generated_at").
			Default(time.Now).
			Comment("AI generation timestamp"),
		field.Int("version").
			Default(1).
			Comment("Mindmap version for regeneration tracking"),
	}
}

func (MindmapGraph) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("session", Session.Type).
			Ref("mindmap").
			Unique().
			Required(),
	}
}
