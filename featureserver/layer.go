package featureserver

type LayerID = uint8

type Layer struct {
	ID LayerID
	fs *FeatureServerClient
}

func (fs *FeatureServerClient) Layer(id LayerID) *Layer {
	l := Layer{id, fs}
	return &l
}
