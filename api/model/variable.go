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
	// MetadataVarPrefix is the prefix of a metadata var name.
	MetadataVarPrefix = "_feature_"
)

// Variable represents a single variable description within a dataset.
type Variable struct {
	Label            string      `json:"label"`
	Key              string      `json:"key"`
	Index            int         `json:"colIndex"`
	Type             string      `json:"type"`
	OriginalType     string      `json:"originalType"` // needed for eval only
	Role             string      `json:"role"`
	DistilRole       string      `json:"distilRole"`
	DisplayVariable  string      `json:"varDisplayName"`
	OriginalVariable string      `json:"varOriginalName"`
	Importance       int         `json:"importance"`
	SuggestedTypes   interface{} `json:"suggestedTypes"`
	Deleted          bool        `json:"deleted"`
}
