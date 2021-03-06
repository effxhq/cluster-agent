package client_plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	effxClusterNameHeader = "X-Effx-Cluster-Name"
)

type Grant struct {
	Name    string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Allowed bool   `protobuf:"varint,2,opt,name=allowed,proto3" json:"allowed,omitempty"`
}

type IntegrationConfig struct {
	Grants []*Grant `protobuf:"bytes,3,rep,name=grants,proto3" json:"grants,omitempty"`
}

type GetResponse struct {
	AccountId         string               `protobuf:"bytes,1,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty"`
	IntegrationName   string               `protobuf:"bytes,2,opt,name=integration_name,json=integrationName,proto3" json:"integration_name,omitempty"`
	IntegrationConfig []*IntegrationConfig `protobuf:"bytes,4,rep,name=integration_config,json=integrationConfig,proto3" json:"integration_config,omitempty"`
}

type httpClient struct {
	BaseURL     string `envconfig:"EFFX_BASE_URL"     default:"https://in.effx.io"`
	ExternalID  string `envconfig:"EFFX_EXTERNAL_ID"`
	SecretKey   string `envconfig:"EFFX_SECRET_KEY"`
	ClusterName string `envconfig:"EFFX_CLUSTER_NAME"`
}

type HTTPClient interface {
	PostResource(ctx context.Context, obj interface{}) error
	FetchConfig(ctx context.Context) (*IntegrationConfig, error)
	IsResourceAllowed(ctx context.Context, requiredGrants ...string) (bool, error)
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
	request, err := json.Marshal(obj)

	if err != nil {
		return errors.Wrap(err, "failed to marshal the request")
	}

	// /v3/hooks/kubernetes/:external_ID
	endpoint := c.BaseURL + "/v3/hooks/kubernetes/" + c.ExternalID

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, bytes.NewBuffer(request))
	if err != nil {
		return errors.Wrap(err, "failed to form request")
	}

	if c, ok := obj.(*metav1.ObjectMeta); ok && c.GetUID() != "" {
		// need uid and version to properly report resources
		requestID := string(c.GetUID()) + "-" + c.GetResourceVersion()
		req.Header.Set("X-Request-ID", requestID)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Token token=%v", c.SecretKey))
	req.Header.Set(effxClusterNameHeader, c.ClusterName)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to post")
	}

	defer resp.Body.Close()

	return nil
}

func (c httpClient) FetchConfig(ctx context.Context) (*IntegrationConfig, error) {
	var (
		getResponse *GetResponse
		result      *IntegrationConfig
	)

	// /v3/integrations/kubernetes/config/:external_id
	endpoint := c.BaseURL + "/v3/integrations/kubernetes/config/" + c.ExternalID

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return result, errors.Wrap(err, "failed to form request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, errors.Wrap(err, "failed to get")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(body, getResponse)
	if err != nil {
		return result, errors.Wrap(err, "failed to unmarshal the response")
	}

	if len(getResponse.IntegrationConfig) == 0 {
		return result, errors.Wrap(err, "no integration config found")
	}

	return getResponse.IntegrationConfig[0], nil
}

func (c httpClient) IsResourceAllowed(ctx context.Context, requiredGrants ...string) (bool, error) {
	resp, err := c.FetchConfig(ctx)
	if err != nil {
		return false, err
	}

	g := make(map[string]bool, len(resp.Grants))
	for _, grant := range resp.Grants {
		g[strings.ToLower(grant.Name)] = grant.Allowed
	}

	for _, requiredGrant := range requiredGrants {
		if !g[strings.ToLower(requiredGrant)] {
			return false, nil
		}
	}

	return false, nil
}
