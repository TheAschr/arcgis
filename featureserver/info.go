package featureserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/mitchellh/mapstructure"
)

const (
	LayerTypeFeatureLayer = "Feature Layer"
	LayerTypeTable        = "Table"
)

// Can be one of:
//   - LayerTypeFeatureLayer
//   - LayerTypeTable
type Info = interface{}

type FieldInfo struct {
	Name string `json:"name"`
	// Can be one of:
	//  - FieldTypeOID
	//  - FieldTypeSmallInt
	//  - FieldTypeInt
	//  - FieldTypeFloat
	//  - FieldTypeDouble
	//  - FieldTypeString
	//  - FieldTypeDate
	Type     string `json:"type"`
	Alias    string `json:"alias"`
	Nullable bool   `json:"nullable"`
}

type FeatureLayerInfo struct {
	ID             LayerID `json:"id"`
	CurrentVersion float32 `json:"currentVersion"`
	Name           string  `json:"name"`
	// Should be LayerTypeFeatureLayer
	Type string `json:"type"`
	// Can be one of:
	//  - GeometryTypePoint
	//  - GeometryTypeMultiPoint
	GeometryType string `json:"geometryType"`
	Fields       []FieldInfo
}

type TableInfo struct {
	ID             LayerID `json:"id"`
	CurrentVersion float32 `json:"currentVersion"`
	Name           string  `json:"name"`
	// Should be LayerTypeTable
	Type   string `json:"type"`
	Fields []FieldInfo
}

func (l *Layer) Info(ctx context.Context) (info Info, err error) {
	u, err := url.Parse(l.fs.url)
	if err != nil {
		return info, fmt.Errorf("failed to parse url: %w", err)
	}

	p, err := url.JoinPath(u.Path, fmt.Sprintf("%d", l.ID))
	if err != nil {
		return info, fmt.Errorf("failed to join path: %w", err)
	}

	u.Path = p

	q := u.Query()
	q.Add("f", "json")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return info, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := l.fs.httpClient.Do(req)
	if err != nil {
		return info, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return info, ErrNotFound
		default:
			return info, fmt.Errorf("unhandled status code: %d", resp.StatusCode)
		}
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return info, fmt.Errorf("failed to read response body: %w", err)
	}

	var respJSON map[string]interface{}
	if err := json.Unmarshal(respBody, &respJSON); err != nil {
		return info, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	respError, ok := respJSON["error"]
	if ok {
		var errRespErr ErrResponseError
		if err := mapstructure.Decode(respError, &errRespErr); err != nil {
			return info, fmt.Errorf("failed to decode error response: %w", err)
		}
		return info, errRespErr
	}

	layerType, ok := respJSON["type"]
	if !ok {
		return info, fmt.Errorf("missing layer type in response")
	}

	switch layerType {
	case LayerTypeFeatureLayer:
		var info FeatureLayerInfo
		if err := mapstructure.Decode(respJSON, &info); err != nil {
			return info, fmt.Errorf("failed to decode feature layer info: %w", err)
		}
		return info, nil
	case LayerTypeTable:
		var info TableInfo
		if err := mapstructure.Decode(respJSON, &info); err != nil {
			return info, fmt.Errorf("failed to decode table info: %w", err)
		}
		return info, nil
	default:
		return info, fmt.Errorf("unhandled layer type: %s", layerType)
	}
}
