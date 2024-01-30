package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/TheAschr/arcgis"
	"github.com/TheAschr/arcgis/featureserver"
)

func main() {
	fsc, err := featureserver.NewClient("https://sampleserver6.arcgisonline.com/arcgis/rest/services/Wildfire/FeatureServer")
	if err != nil {
		log.Fatalf("failed to create feature server client: %v", err)
	}

	layerID := uint8(0)

	lInfo, err := fsc.Layer(layerID).Info(context.Background())
	if err != nil {
		log.Fatalf("failed to get layer info: %v", err)
	}

	// Expect FeatureLayerInfo
	flInfo := lInfo.(featureserver.FeatureLayerInfo)

	type NewFeatureAttributes struct {
		Rotation *int16 `json:"rotation"`
	}

	type NewFeature struct {
		Attributes NewFeatureAttributes        `json:"attributes"`
		Geometry   featureserver.GeometryPoint `json:"geometry"`
	}

	if err := featureserver.ValidateFeature(NewFeature{}, flInfo); err != nil {
		log.Fatalf("failed to validate feature: %v", err)
	}

	newFeature := NewFeature{
		Attributes: NewFeatureAttributes{
			Rotation: arcgis.Nullable(int16(-1)),
		},
		Geometry: featureserver.GeometryPoint{
			X: -122.6764,
			Y: 45.5165,
		},
	}

	type UpdateFeatureAttributes struct {
		ObjectID int32  `json:"objectid"`
		Rotation *int16 `json:"rotation"`
	}

	type UpdateFeature struct {
		Attributes UpdateFeatureAttributes     `json:"attributes"`
		Geometry   featureserver.GeometryPoint `json:"geometry"`
	}

	if err := featureserver.ValidateFeature(UpdateFeature{}, flInfo); err != nil {
		log.Fatalf("failed to validate feature: %v", err)
	}

	updateFeature := UpdateFeature{
		Attributes: UpdateFeatureAttributes{
			ObjectID: 7282114,
			Rotation: arcgis.Nullable(int16(-2)),
		},
		Geometry: featureserver.GeometryPoint{
			X: -122.6764,
			Y: 45.5165,
		},
	}

	edits := []featureserver.Edit{
		{
			LayerID: flInfo.ID,
			Adds:    []featureserver.AddOperation{newFeature},
			Updates: []featureserver.UpdateOperation{updateFeature},
			Deletes: []featureserver.DeleteOperation{7282115},
		},
	}

	log.Printf("Applying edits to layer name: '%s'", flInfo.Name)
	res, err := fsc.Layer(flInfo.ID).ApplyEdits(context.Background(), featureserver.ApplyEditsVariables{
		Edits: edits,
	})
	if err != nil {
		log.Fatalf("failed to apply edits: %v", err)
	}

	j, _ := json.MarshalIndent(res, "", "  ")

	log.Printf("Success: %s", j)
}
