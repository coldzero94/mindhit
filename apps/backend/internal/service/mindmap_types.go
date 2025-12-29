package service

// MindmapNode represents a node in the mindmap graph.
type MindmapNode struct {
	ID       string                 `json:"id"`
	Label    string                 `json:"label"`
	Type     string                 `json:"type"` // core, topic, subtopic, page
	Size     float64                `json:"size"`
	Color    string                 `json:"color"`
	Position *Position              `json:"position,omitempty"`
	Data     map[string]interface{} `json:"data"`
}

// Position represents 3D coordinates.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// MindmapEdge represents a connection between nodes.
type MindmapEdge struct {
	Source string  `json:"source"`
	Target string  `json:"target"`
	Weight float64 `json:"weight"`
	Label  string  `json:"label,omitempty"`
}

// MindmapLayout defines the layout configuration.
type MindmapLayout struct {
	Type   string                 `json:"type"` // galaxy, tree, radial
	Params map[string]interface{} `json:"params"`
}

// MindmapData contains the complete mindmap structure.
type MindmapData struct {
	Nodes  []MindmapNode `json:"nodes"`
	Edges  []MindmapEdge `json:"edges"`
	Layout MindmapLayout `json:"layout"`
}

// ConvertNodesToMaps converts MindmapNode slice to []map[string]interface{} for Ent storage.
func ConvertNodesToMaps(nodes []MindmapNode) []map[string]interface{} {
	result := make([]map[string]interface{}, len(nodes))
	for i, node := range nodes {
		m := map[string]interface{}{
			"id":    node.ID,
			"label": node.Label,
			"type":  node.Type,
			"size":  node.Size,
			"color": node.Color,
			"data":  node.Data,
		}
		if node.Position != nil {
			m["position"] = map[string]interface{}{
				"x": node.Position.X,
				"y": node.Position.Y,
				"z": node.Position.Z,
			}
		}
		result[i] = m
	}
	return result
}

// ConvertEdgesToMaps converts MindmapEdge slice to []map[string]interface{} for Ent storage.
func ConvertEdgesToMaps(edges []MindmapEdge) []map[string]interface{} {
	result := make([]map[string]interface{}, len(edges))
	for i, edge := range edges {
		m := map[string]interface{}{
			"source": edge.Source,
			"target": edge.Target,
			"weight": edge.Weight,
		}
		if edge.Label != "" {
			m["label"] = edge.Label
		}
		result[i] = m
	}
	return result
}

// ConvertLayoutToMap converts MindmapLayout to map[string]interface{} for Ent storage.
func ConvertLayoutToMap(layout MindmapLayout) map[string]interface{} {
	return map[string]interface{}{
		"type":   layout.Type,
		"params": layout.Params,
	}
}
