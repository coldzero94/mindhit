package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Highlight struct {
	ent.Schema
}

func (Highlight) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (Highlight) Fields() []ent.Field {
	return []ent.Field{
		field.Text("text").
			NotEmpty().
			Comment("Highlighted text content"),
		field.String("selector").
			Optional().
			Comment("CSS selector for highlight position"),
		field.String("color").
			Default("#FFFF00").
			Comment("Highlight color (hex)"),
		field.String("note").
			Optional().
			Comment("User note for this highlight"),
	}
}

func (Highlight) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("session", Session.Type).
			Ref("highlights").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("page_visit", PageVisit.Type).
			Unique(),
	}
}
