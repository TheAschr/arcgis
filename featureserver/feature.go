package featureserver

import (
	"encoding/json"
)

type Feature struct {
	Attributes json.RawMessage `json:"attributes"`
	// Can be one of:
	//  - GeometryNone
	//  - GeometryPoint
	//  - GeometryMultiPoint
	Geometry interface{} `json:"geometry"`
}
