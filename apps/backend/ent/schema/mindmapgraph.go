package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// MindmapGraph represents a mindmap visualization for a session.
type MindmapGraph struct {
	ent.Schema
}

// Mixin returns the mixins for MindmapGraph.
func (MindmapGraph) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Fields returns the fields for MindmapGraph.
func (MindmapGraph) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("status").
			Values("pending", "generating", "completed", "failed").
			Default("pending").
			Comment("Mindmap generation status"),
		field.JSON("nodes", []map[string]interface{}{}).
			Optional().
			Comment("Mindmap node data"),
		field.JSON("graph_edges", []map[string]interface{}{}).
			Optional().
			Comment("Mindmap edge data"),
		field.JSON("layout", map[string]interface{}{}).
			Optional().
			Comment("Layout configuration"),
		field.String("error_message").
			Optional().
			Nillable().
			Comment("Error message if generation failed"),
		field.Time("generated_at").
			Default(time.Now).
			Comment("AI generation timestamp"),
		field.Int("version").
			Default(1).
			Comment("Mindmap version for regeneration tracking"),
	}
}

// Edges returns the edges for MindmapGraph.
func (MindmapGraph) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("session", Session.Type).
			Ref("mindmap").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
