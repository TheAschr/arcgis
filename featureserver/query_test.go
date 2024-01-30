package featureserver

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/TheAschr/arcgis"
)

func TestLayerQuery(t *testing.T) {
	t.Run("Results", func(t *testing.T) {

		type FieldTest struct {
			Name   string
			Alias  string
			Type   string
			Length int
		}

		type LayerTest struct {
			ID                   LayerID
			ObjectIDFieldName    string
			GlobalIDFieldName    string
			GeometryType         string
			Fields               []FieldTest
			ValidateAttributesFn func(f Feature)
		}

		type FeatureServerTest struct {
			URL    string
			Layers []LayerTest
		}

		fsTests := []FeatureServerTest{
			{
				URL: "https://sampleserver6.arcgisonline.com/arcgis/rest/services/Wildfire/FeatureServer",
				Layers: []LayerTest{
					{
						ID:                0,
						ObjectIDFieldName: "objectid",
						GlobalIDFieldName: "",
						GeometryType:      "esriGeometryPoint",
						Fields: []FieldTest{
							{
								Name:   "objectid",
								Alias:  "OBJECTID",
								Type:   FieldTypeOID,
								Length: 0,
							},
							{
								Name:   "rotation",
								Alias:  "Rotation",
								Type:   FieldTypeSmallInt,
								Length: 0,
							},
							{
								Name:   "description",
								Alias:  "Description",
								Type:   FieldTypeString,
								Length: 75,
							},
							{
								Name:   "eventdate",
								Alias:  "Date",
								Type:   FieldTypeDate,
								Length: 8,
							},
							{
								Name:   "eventtype",
								Alias:  "Type",
								Type:   FieldTypeInt,
								Length: 0,
							},
							{
								Name:   "created_user",
								Alias:  "created_user",
								Type:   FieldTypeString,
								Length: 255,
							},
							{
								Name:   "created_date",
								Alias:  "created_date",
								Type:   FieldTypeDate,
								Length: 8,
							},
							{
								Name:   "last_edited_user",
								Alias:  "last_edited_user",
								Type:   FieldTypeString,
								Length: 255,
							},
							{
								Name:   "last_edited_date",
								Alias:  "last_edited_date",
								Type:   FieldTypeDate,
								Length: 8,
							},
						},
						ValidateAttributesFn: func(f Feature) {
							type Attributes struct {
								ObjectID       int          `json:"objectid"`
								Rotation       *int         `json:"rotation"`
								Description    *string      `json:"description"`
								EventDate      *arcgis.Date `json:"eventdate"`
								EventType      int          `json:"eventtype"`
								CreatedUser    string       `json:"created_user"`
								CreatedDate    arcgis.Date  `json:"created_date"`
								LastEditedUser string       `json:"last_edited_user"`
								LastEditedDate arcgis.Date  `json:"last_edited_date"`
							}

							var attributes Attributes
							if err := json.Unmarshal(f.Attributes, &attributes); err != nil {
								t.Fatalf("failed to unmarshal attributes: %v", err)
							}

							attributeTests := map[int]Attributes{
								4247677: {
									ObjectID:       4247677,
									Rotation:       nil,
									Description:    nil,
									EventDate:      nil,
									EventType:      12,
									CreatedUser:    "USER1",
									CreatedDate:    arcgis.NewDateFromUnixMillis(1706498064000),
									LastEditedUser: "USER1",
									LastEditedDate: arcgis.NewDateFromUnixMillis(1706498064000),
								},
								4247698: {
									ObjectID:       4247698,
									Rotation:       arcgis.Nullable(54),
									Description:    arcgis.Nullable("gvfyghvjnm"),
									EventDate:      arcgis.Nullable(arcgis.NewDateFromUnixMillis(1706557571000)),
									EventType:      1,
									CreatedUser:    "",
									CreatedDate:    arcgis.NewDateFromUnixMillis(1706557579000),
									LastEditedUser: "",
									LastEditedDate: arcgis.NewDateFromUnixMillis(1706557579000),
								},
							}

							attributeTest, ok := attributeTests[attributes.ObjectID]
							if !ok {
								return
							}

							if attributeTest.ObjectID != attributes.ObjectID {
								t.Errorf("expected object id %d, got: %d", attributeTest.ObjectID, attributes.ObjectID)
							}

							if attributeTest.Rotation != attributes.Rotation {
								t.Errorf("expected rotation %v, got: %v", attributeTest.Rotation, attributes.Rotation)
							}

							if attributeTest.Description != attributes.Description {
								t.Errorf("expected description %v, got: %v", attributeTest.Description, attributes.Description)
							}

							if attributeTest.EventDate != attributes.EventDate {
								t.Errorf("expected event date %v, got: %v", attributeTest.EventDate, attributes.EventDate)
							}

							if attributeTest.EventType != attributes.EventType {
								t.Errorf("expected event type %d, got: %d", attributeTest.EventType, attributes.EventType)
							}

							if attributeTest.CreatedUser != attributes.CreatedUser {
								t.Errorf("expected created user '%s', got: '%s'", attributeTest.CreatedUser, attributes.CreatedUser)
							}

							if !attributeTest.CreatedDate.Equal(*attributes.CreatedDate.Time) {
								t.Errorf("expected created date %v, got: %v", attributeTest.CreatedDate, attributes.CreatedDate)
							}

							if attributeTest.LastEditedUser != attributes.LastEditedUser {
								t.Errorf("expected last edited user '%s', got: '%s'", attributeTest.LastEditedUser, attributes.LastEditedUser)
							}

							if !attributeTest.LastEditedDate.Equal(*attributes.LastEditedDate.Time) {
								t.Errorf("expected last edited date %v, got: %v", attributeTest.LastEditedDate, attributes.LastEditedDate)
							}
						},
					},
				},
			},
		}

		for _, fsTest := range fsTests {
			fsc, err := NewClient(fsTest.URL)
			if err != nil {
				t.Fatalf("failed to create feature server client: %v", err)
			}

			for _, layerTest := range fsTest.Layers {
				outFields := make([]string, len(layerTest.Fields))

				for i, field := range layerTest.Fields {
					outFields[i] = field.Name
				}

				results, err := fsc.Layer(layerTest.ID).Query(context.Background(), QueryVariables{
					Where:          "1=1",
					OutFields:      outFields,
					ReturnGeometry: true,
				})
				if err != nil {
					t.Fatalf("failed to query layer: %v", err)
				}

				if results.ObjectIDFieldName != layerTest.ObjectIDFieldName {
					t.Errorf("expected object id field name %s, got: %s", layerTest.ObjectIDFieldName, results.ObjectIDFieldName)
				}

				if results.GlobalIDFieldName != layerTest.GlobalIDFieldName {
					t.Errorf("expected global id field name %s, got: %s", layerTest.GlobalIDFieldName, results.GlobalIDFieldName)
				}

				if results.GeometryType != layerTest.GeometryType {
					t.Errorf("expected geometry type %s, got: %s", layerTest.GeometryType, results.GeometryType)
				}

				if len(results.Fields) != len(layerTest.Fields) {
					t.Errorf("expected %d fields, got: %d", len(layerTest.Fields), len(results.Fields))
				}

				for i, field := range results.Fields {
					if field.Name != layerTest.Fields[i].Name {
						t.Errorf("expected field name %s, got: %s", layerTest.Fields[i].Name, field.Name)
					}

					if field.Alias != layerTest.Fields[i].Alias {
						t.Errorf("expected field alias %s, got: %s", layerTest.Fields[i].Alias, field.Alias)
					}

					if field.Type != layerTest.Fields[i].Type {
						t.Errorf("expected field type %s, got: %s", layerTest.Fields[i].Type, field.Type)
					}

					if field.Length != layerTest.Fields[i].Length {
						t.Errorf("expected field length %d, got: %d", layerTest.Fields[i].Length, field.Length)
					}
				}

				if len(results.Features) == 0 {
					t.Errorf("expected features, got none")
				} else {
					f := results.Features[0]

					switch g := f.Geometry.(type) {
					case GeometryPoint:
						if layerTest.GeometryType != GeometryTypePoint {
							t.Errorf("expected geometry type %s, got: %s", layerTest.GeometryType, GeometryTypePoint)
						}
						if g.X == 0 {
							t.Errorf("expected x coordinate, got none")
						}
						if g.Y == 0 {
							t.Errorf("expected y coordinate, got none")
						}
					case GeometryMultiPoint:
						if layerTest.GeometryType != GeometryTypeMultiPoint {
							t.Errorf("expected geometry type %s, got: %s", layerTest.GeometryType, GeometryTypeMultiPoint)
						}

						if len(g.Points) == 0 {
							t.Errorf("expected multipoint, got none")
						}
					default:
						t.Errorf("unhandled geometry type: %T", g)
					}

					if layerTest.ValidateAttributesFn != nil {
						layerTest.ValidateAttributesFn(f)
					}
				}
			}
		}

	})
}
