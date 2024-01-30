package featureserver

const (
	GeometryTypeNone       = ""
	GeometryTypePoint      = "esriGeometryPoint"
	GeometryTypeMultiPoint = "esriGeometryMultipoint"
)

type GeometryNone struct{}

type GeometryMultiPoint struct {
	Points [][]float64 `json:"points"`
}

type GeometryPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
