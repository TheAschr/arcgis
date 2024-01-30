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

	"github.com/mitchellh/mapstructure"
)

type AddOperation = interface{}
type UpdateOperation = interface{}
type DeleteOperation = interface{}

type Edit struct {
	LayerID uint8             `json:"id"`
	Adds    []AddOperation    `json:"adds,omitempty"`
	Updates []UpdateOperation `json:"updates,omitempty"`
	Deletes []DeleteOperation `json:"deletes,omitempty"`
}

type ApplyEditsVariables struct {
	Edits []Edit `json:"edits"`
}

type LayerEditResults struct {
	LayerID       uint8             `json:"id"`
	AddResults    []json.RawMessage `json:"addsResults,omitempty"`
	UpdateResults []json.RawMessage `json:"updateResults,omitempty"`
	DeleteResults []json.RawMessage `json:"deleteResults,omitempty"`
}

type ApplyEditsResults = []LayerEditResults

func (l *Layer) ApplyEdits(ctx context.Context, variables ApplyEditsVariables) (results ApplyEditsResults, err error) {
	u, err := url.Parse(l.fs.url)
	if err != nil {
		return results, fmt.Errorf("failed to parse url: %w", err)
	}

	p, err := url.JoinPath(u.Path, "applyEdits")
	if err != nil {
		return results, fmt.Errorf("failed to join path: %w", err)
	}

	u.Path = p

	formBody := &bytes.Buffer{}
	formBodyWriter := multipart.NewWriter(formBody)

	editsJSON, err := json.Marshal(variables.Edits)
	if err != nil {
		return results, fmt.Errorf("failed to marshal 'edits' field: %w", err)
	}

	if err := formBodyWriter.WriteField("edits", string(editsJSON)); err != nil {
		return results, fmt.Errorf("failed to write 'edits' field: %w", err)
	}

	if err := formBodyWriter.WriteField("f", "json"); err != nil {
		return results, fmt.Errorf("failed to write 'f' field: %w", err)
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

	var respErr map[string]interface{}
	if err := json.Unmarshal(respBody, &respErr); err == nil {
		respError, ok := respErr["error"]
		if ok {
			var errRespErr ErrResponseError
			if err := mapstructure.Decode(respError, &errRespErr); err != nil {
				return results, fmt.Errorf("failed to decode error response: %w", err)
			}
			return results, errRespErr
		}
	}

	if err := json.Unmarshal(respBody, &results); err != nil {
		return results, fmt.Errorf("failed to decode query results: %w", err)
	}

	return results, nil
}
