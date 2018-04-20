package model

const (
	// FeatureTypeTrain is the training feature type.
	FeatureTypeTrain = "train"
	// FeatureTypeTarget is the target feature type.
	FeatureTypeTarget = "target"
	// RoleIndex is the role used for index fields.
	RoleIndex = "index"
	// D3MIndexFieldName denotes the name of the index field.
	D3MIndexFieldName = "d3mIndex"
)

// Variable represents a single variable description within a dataset.
type Variable struct {
	Name             string      `json:"name"`
	Type             string      `json:"type"`
	Role             string      `json:"role"`
	DisplayVariable  string      `json:"varDisplayName"`
	OriginalVariable string      `json:"varOriginalName"`
	Importance       int         `json:"importance"`
	SuggestedTypes   interface{} `json:"suggestedTypes"`
}
