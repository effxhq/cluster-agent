package client_plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type httpClient struct {
	BaseURL    string `envconfig:"EFFX_BASE_URL"`
	ExternalID string `envconfig:"EFFX_EXTERNAL_ID"`  // uuid
	SecretKey  string `envconfig:"EFFX_SECRET_KEY"`   // uuid
}

type HTTPClient interface {
	PostResource(ctx context.Context, obj interface{}) error
	FetchConfig(ctx context.Context) error
}

func NewHTTPClient() (HTTPClient, error) {
	httpClient := &httpClient{}

	err := envconfig.Process("", httpClient)
	if err != nil {
		return httpClient, errors.Wrap(err, "failed to read config from environment")
	}

	return httpClient, nil
}

func (c httpClient) PostResource(ctx context.Context, obj interface{}) error {
	// /v3/hooks/kubernetes/:external_ID
	request, err := json.Marshal(obj)

	if err != nil {
		return errors.Wrap(err, "failed to marshal the request")
	}

	resp, err := http.Post(c.BaseURL+"/v3/hooks/kubernetes/"+c.ExternalID, "application/json", bytes.NewBuffer(request))
	if err != nil {
		return errors.Wrap(err, "failed to post")
	}

	defer resp.Body.Close()

	return nil
}

func (c httpClient) FetchConfig(ctx context.Context) error {
	// /v3/integrations/kubernetes/config/:external_id

	// TODO: fetch grants and check these before posting
	return status.Errorf(codes.Unimplemented, "fetch config not implemented")
}
