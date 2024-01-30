package main

import (
	"context"
	"encoding/json"
	"fmt"
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

	log.Printf("Querying layer name: '%s'", flInfo.Name)
	res, err := fsc.Layer(flInfo.ID).Query(context.Background(), featureserver.QueryVariables{
		Where: "1=1",
		OutFields: []string{
			"objectid",
			"rotation",
			"description",
			"eventdate",
			"eventtype",
			"created_user",
			"created_date",
			"last_edited_user",
			"last_edited_date",
		},
		ReturnGeometry: false,
	})
	if err != nil {
		log.Fatalf("failed to query layer: %v", err)
	}

	type FeatureAttributes struct {
		ObjectID       int32        `json:"objectid"`
		Rotation       *int16       `json:"rotation"`
		Description    *string      `json:"description"`
		EventDate      *arcgis.Date `json:"eventdate"`
		EventType      *int32       `json:"eventtype"`
		CreatedUser    *string      `json:"created_user"`
		CreatedDate    *arcgis.Date `json:"created_date"`
		LastEditedUser *string      `json:"last_edited_user"`
		LastEditedDate *arcgis.Date `json:"last_edited_date"`
	}

	// Validate attributes struct against layer info
	if err := featureserver.ValidateFeature(FeatureAttributes{}, flInfo); err != nil {
		log.Fatalf("failed to validate attributes: %v", err)
	}

	for _, f := range res.Features {
		var attributes FeatureAttributes
		if err := json.Unmarshal(f.Attributes, &attributes); err != nil {
			log.Fatalf("failed to unmarshal attributes: %v", err)
		}

		fmt.Printf(`
ObjectID: %d
Rotation: %s
Description: %s
EventDate: %s
EventType: %s
CreatedUser: %s
CreatedDate: %s
LastEditedUser: %s
LastEditedDate: %s
`,
			attributes.ObjectID,
			fmtPointer(attributes.Rotation),
			fmtPointer(attributes.Description),
			fmtPointer(attributes.EventDate),
			fmtPointer(attributes.EventType),
			fmtPointer(attributes.CreatedUser),
			fmtPointer(attributes.CreatedDate),
			fmtPointer(attributes.LastEditedUser),
			fmtPointer(attributes.LastEditedDate),
		)
	}
}

func fmtPointer[T any](p *T) string {
	if p == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", *p)
}
