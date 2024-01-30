package arcgis

func Nullable[T any](v T) *T {
	return &v
}
