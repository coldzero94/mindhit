package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type URL struct {
	ent.Schema
}

func (URL) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (URL) Fields() []ent.Field {
	return []ent.Field{
		field.String("url").
			NotEmpty().
			Comment("Original URL"),
		field.String("url_hash").
			Unique().
			NotEmpty().
			Comment("SHA256 hash of normalized URL"),
		field.String("title").
			Optional().
			Comment("Page title"),
		field.Text("content").
			Optional().
			Comment("Extracted page content"),
		field.Text("summary").
			Optional().
			Comment("AI-generated summary"),
		field.JSON("keywords", []string{}).
			Optional().
			Comment("AI-extracted keywords"),
		field.Time("crawled_at").
			Optional().
			Nillable().
			Comment("Last time the URL content was crawled"),
	}
}

func (URL) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("page_visits", PageVisit.Type).
			Ref("url"),
	}
}

func (URL) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("url_hash"),
	}
}
