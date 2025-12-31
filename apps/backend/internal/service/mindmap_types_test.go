package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertNodesToMaps(t *testing.T) {
	nodes := []MindmapNode{
		{
			ID:    "core",
			Label: "Core Theme",
			Type:  "core",
			Size:  100,
			Color: "#FFD700",
			Position: &Position{
				X: 0,
				Y: 0,
				Z: 0,
			},
			Data: map[string]interface{}{
				"description": "Test description",
			},
		},
		{
			ID:    "topic-1",
			Label: "Topic 1",
			Type:  "topic",
			Size:  50,
			Color: "#3B82F6",
			Data: map[string]interface{}{
				"keywords": []string{"key1", "key2"},
			},
		},
	}

	result := ConvertNodesToMaps(nodes)

	assert.Len(t, result, 2)

	// Check first node (with position)
	assert.Equal(t, "core", result[0]["id"])
	assert.Equal(t, "Core Theme", result[0]["label"])
	assert.Equal(t, "core", result[0]["type"])
	assert.Equal(t, 100.0, result[0]["size"])
	assert.Equal(t, "#FFD700", result[0]["color"])
	assert.NotNil(t, result[0]["position"])

	pos := result[0]["position"].(map[string]interface{})
	assert.Equal(t, 0.0, pos["x"])
	assert.Equal(t, 0.0, pos["y"])
	assert.Equal(t, 0.0, pos["z"])

	// Check second node (without position)
	assert.Equal(t, "topic-1", result[1]["id"])
	assert.Nil(t, result[1]["position"])
}

func TestConvertNodesToMaps_Empty(t *testing.T) {
	result := ConvertNodesToMaps([]MindmapNode{})
	assert.Empty(t, result)
}

func TestConvertEdgesToMaps(t *testing.T) {
	edges := []MindmapEdge{
		{
			Source: "core",
			Target: "topic-1",
			Weight: 1.0,
		},
		{
			Source: "topic-1",
			Target: "topic-2",
			Weight: 0.5,
			Label:  "related topics",
		},
	}

	result := ConvertEdgesToMaps(edges)

	assert.Len(t, result, 2)

	// Check first edge (without label)
	assert.Equal(t, "core", result[0]["source"])
	assert.Equal(t, "topic-1", result[0]["target"])
	assert.Equal(t, 1.0, result[0]["weight"])
	_, hasLabel := result[0]["label"]
	assert.False(t, hasLabel)

	// Check second edge (with label)
	assert.Equal(t, "topic-1", result[1]["source"])
	assert.Equal(t, "topic-2", result[1]["target"])
	assert.Equal(t, 0.5, result[1]["weight"])
	assert.Equal(t, "related topics", result[1]["label"])
}

func TestConvertEdgesToMaps_Empty(t *testing.T) {
	result := ConvertEdgesToMaps([]MindmapEdge{})
	assert.Empty(t, result)
}

func TestConvertLayoutToMap(t *testing.T) {
	layout := MindmapLayout{
		Type: "galaxy",
		Params: map[string]interface{}{
			"center": []float64{0, 0, 0},
			"scale":  1.0,
		},
	}

	result := ConvertLayoutToMap(layout)

	assert.Equal(t, "galaxy", result["type"])
	assert.NotNil(t, result["params"])

	params := result["params"].(map[string]interface{})
	assert.Equal(t, 1.0, params["scale"])
}

func TestConvertLayoutToMap_EmptyParams(t *testing.T) {
	layout := MindmapLayout{
		Type:   "tree",
		Params: nil,
	}

	result := ConvertLayoutToMap(layout)

	assert.Equal(t, "tree", result["type"])
	assert.Nil(t, result["params"])
}
