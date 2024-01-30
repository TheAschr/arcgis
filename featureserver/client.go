package featureserver

import "net/http"

type ClientOption = func(*FeatureServerClient) error

func WithHttpClient(c http.Client) ClientOption {
	return func(fs *FeatureServerClient) error {
		fs.httpClient = c
		return nil
	}
}

type FeatureServerClient struct {
	url        string
	httpClient http.Client
}

func NewClient(url string, opts ...ClientOption) (*FeatureServerClient, error) {
	fs := FeatureServerClient{
		url:        url,
		httpClient: http.Client{},
	}

	for _, opt := range opts {
		if err := opt(&fs); err != nil {
			return nil, err
		}
	}

	return &fs, nil
}
