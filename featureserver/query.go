package featureserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type QueryVariables struct {
	// A SQL where clause for the query filter. Any legal SQL where clause operating on the fields in the layer is allowed.
	Where string
	// If true, the result set includes the geometry associated with each feature. The default is false.
	ReturnGeometry bool
	// A comma delimited list of field names. If you specify the shape field in the list of return fields, it is ignored. To request geometry, set returnGeometry to true.
	OutFields []string
	// Limits the number of features returned by a query to a specified number.
	ResultRecordCount int
}

type QueryResults struct {
	ObjectIDFieldName string `json:"objectIdFieldName"`
	GlobalIDFieldName string `json:"globalIdFieldName"`
	// Can be one of:
	//  - GeometryTypeNone
	//  - GeometryTypePoint
	//  - GeometryTypeMultiPoint
	GeometryType string    `json:"geometryType"`
	Fields       []Field   `json:"fields"`
	Features     []Feature `json:"features"`
}

func (l *Layer) Query(ctx context.Context, variables QueryVariables) (results QueryResults, err error) {
	u, err := url.Parse(l.fs.url)
	if err != nil {
		return results, fmt.Errorf("failed to parse url: %w", err)
	}

	p, err := url.JoinPath(u.Path, fmt.Sprintf("%d", l.ID), "query")
	if err != nil {
		return results, fmt.Errorf("failed to join path: %w", err)
	}

	u.Path = p

	formBody := &bytes.Buffer{}
	formBodyWriter := multipart.NewWriter(formBody)

	if err := formBodyWriter.WriteField("where", variables.Where); err != nil {
		return results, fmt.Errorf("failed to write 'where' field: %w", err)
	}

	if err := formBodyWriter.WriteField("f", "json"); err != nil {
		return results, fmt.Errorf("failed to write 'f' field: %w", err)
	}

	if err := formBodyWriter.WriteField("returnGeometry", fmt.Sprintf("%t", variables.ReturnGeometry)); err != nil {
		return results, fmt.Errorf("failed to write 'returnGeometry' field: %w", err)
	}

	if variables.OutFields != nil {
		if err := formBodyWriter.WriteField("outFields", strings.Join(variables.OutFields, ",")); err != nil {
			return results, fmt.Errorf("failed to write 'outFields' field: %w", err)
		}
	}

	if variables.ResultRecordCount != -1 {
		if err := formBodyWriter.WriteField("resultRecordCount", fmt.Sprintf("%d", variables.ResultRecordCount)); err != nil {
			return results, fmt.Errorf("failed to write 'resultRecordCount' field: %w", err)
		}
	}

	if err := formBodyWriter.Close(); err != nil {
		return results, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), formBody)
	if err != nil {
		return results, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Content-Type", formBodyWriter.FormDataContentType())

	resp, err := l.fs.httpClient.Do(req)
	if err != nil {
		return results, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return results, ErrNotFound
		default:
			return results, fmt.Errorf("unhandled status code: %d", resp.StatusCode)
		}
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return results, fmt.Errorf("failed to read response body: %w", err)
	}

	var respJSON map[string]interface{}
	if err := json.Unmarshal(respBody, &respJSON); err != nil {
		return results, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	respError, ok := respJSON["error"]
	if ok {
		var errRespErr ErrResponseError
		if err := mapstructure.Decode(respError, &errRespErr); err != nil {
			return results, fmt.Errorf("failed to decode error response: %w", err)
		}
		return results, errRespErr
	}

	if err := json.Unmarshal(respBody, &results); err != nil {
		return results, fmt.Errorf("failed to decode query results: %w", err)
	}

	for i, f := range results.Features {
		if !variables.ReturnGeometry {
			results.Features[i].Geometry = GeometryNone{}
			continue
		}
		switch results.GeometryType {
		case GeometryTypePoint:
			var geometry GeometryPoint
			if err := mapstructure.Decode(f.Geometry, &geometry); err != nil {
				return results, fmt.Errorf("failed to decode point geometry: %w", err)
			}
			results.Features[i].Geometry = geometry
		case GeometryTypeMultiPoint:
			var geometry GeometryMultiPoint
			if err := mapstructure.Decode(f.Geometry, &geometry); err != nil {
				return results, fmt.Errorf("failed to decode multipoint geometry: %w", err)
			}
			results.Features[i].Geometry = geometry
		default:
			return results, fmt.Errorf("unhandled geometry type: %s", results.GeometryType)
		}
	}

	return results, nil
}
