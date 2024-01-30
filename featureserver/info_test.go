package featureserver

import (
	"context"
	"net/http"
	"testing"
)

func TestLayerInfo(t *testing.T) {

	t.Run("Layer not found", func(t *testing.T) {
		type FeatureServerTest struct {
			URL                 string
			NonExistentLayerIDs []LayerID
			ExistentLayerIDs    []LayerID
		}

		fsTests := []FeatureServerTest{
			{
				URL: "https://sampleserver6.arcgisonline.com/arcgis/rest/services/Wildfire/FeatureServer",
				NonExistentLayerIDs: []LayerID{
					50,
					51,
					52,
				},
				ExistentLayerIDs: []LayerID{
					0,
					1,
					2,
				},
			},
		}

		for _, fsTest := range fsTests {
			fsc, err := NewClient(fsTest.URL)
			if err != nil {
				t.Fatalf("failed to create feature server client: %v", err)
			}

			for _, layerID := range fsTest.ExistentLayerIDs {
				if _, err := fsc.Layer(layerID).Info(context.Background()); err != nil {
					t.Errorf("failed to get layer info: %v", err)
				}
			}

			for _, layerID := range fsTest.NonExistentLayerIDs {
				_, err := fsc.Layer(layerID).Info(context.Background())
				switch err := err.(type) {
				case ErrResponseError:
					if err.Code != http.StatusInternalServerError {
						t.Errorf("expected status code %d, got: %d", http.StatusInternalServerError, err.Code)
					}
					if err.Message != "json" {
						t.Errorf("expected message 'json', got: %s", err.Message)
					}
					if len(err.Details) != 0 {
						t.Errorf("expected details to be empty, got: %v", err.Details)
					}
				default:
					t.Errorf("expected ErrResponseError, got: %v", err)
				}
			}
		}
	})

	t.Run("Layer info contents", func(t *testing.T) {
		type LayerTest struct {
			ID             LayerID
			Name           string
			Type           string
			CurrentVersion float32
		}

		type FeatureServerTest struct {
			URL                string
			NonExistentLayerID LayerID
			Layers             []LayerTest
		}

		fsTests := []FeatureServerTest{
			{
				URL:                "https://sampleserver6.arcgisonline.com/arcgis/rest/services/Wildfire/FeatureServer",
				NonExistentLayerID: 50,
				Layers: []LayerTest{
					{
						ID:             0,
						Name:           "Wildfire Response Points",
						Type:           LayerTypeFeatureLayer,
						CurrentVersion: 10.91,
					},
					{
						ID:             1,
						Name:           "Wildfire Response Lines",
						Type:           LayerTypeFeatureLayer,
						CurrentVersion: 10.91,
					},
					{
						ID:             2,
						Name:           "Wildfire Response Polygons",
						Type:           LayerTypeFeatureLayer,
						CurrentVersion: 10.91,
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
				info, err := fsc.Layer(layerTest.ID).Info(context.Background())
				if err != nil {
					t.Fatalf("failed to get layer info: %v", err)
				}

				switch info := info.(type) {
				case FeatureLayerInfo:
					if layerTest.Type != LayerTypeFeatureLayer {
						t.Errorf("expected layer type %s, got: %s", LayerTypeFeatureLayer, layerTest.Type)
					}
					if info.Name != layerTest.Name {
						t.Errorf("expected layer name %s, got: %s", layerTest.Name, info.Name)
					}
					if info.CurrentVersion != layerTest.CurrentVersion {
						t.Errorf("expected layer current version %f, got: %f", layerTest.CurrentVersion, info.CurrentVersion)
					}
				case TableInfo:
					if layerTest.Type != LayerTypeTable {
						t.Errorf("expected layer type %s, got: %s", LayerTypeTable, layerTest.Type)
					}
					if info.Name != layerTest.Name {
						t.Errorf("expected layer name %s, got: %s", layerTest.Name, info.Name)
					}
					if info.CurrentVersion != layerTest.CurrentVersion {
						t.Errorf("expected layer current version %f, got: %f", layerTest.CurrentVersion, info.CurrentVersion)
					}
				default:
					t.Fatalf("info is not FeatureLayerInfo or TableInfo")
				}
			}
		}
	})

}
