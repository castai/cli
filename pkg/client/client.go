package client

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/castai/cli/pkg/client/sdk"
	"github.com/castai/cli/pkg/config"
)

const (
	defaultTimeout = 1 * time.Minute
)

type Interface interface {
	CreateNewCluster(ctx context.Context, req sdk.CreateNewClusterJSONRequestBody) (*sdk.KubernetesCluster, error)
	GetCluster(ctx context.Context, req sdk.ClusterId) (*sdk.KubernetesCluster, error)
	DeleteCluster(ctx context.Context, req sdk.ClusterId) error
	ListRegions(ctx context.Context) ([]sdk.CastRegion, error)
	ListCloudCredentials(ctx context.Context) ([]sdk.CloudCredentials, error)
	GetClusterKubeconfig(ctx context.Context, req sdk.ClusterId) ([]byte, error)
	ListKubernetesClusters(ctx context.Context, req *sdk.ListKubernetesClustersParams) ([]sdk.KubernetesCluster, error)
	ListClusterNodes(ctx context.Context, req sdk.ClusterId) ([]sdk.Node, error)
	ListAuthTokens(ctx context.Context) ([]sdk.AuthToken, error)
	FeedbackEvents(ctx context.Context, req sdk.ClusterId) ([]sdk.KubernetesClusterFeedbackEvent, error)
	SetupNodeSSH(ctx context.Context, clusterID sdk.ClusterId, nodeID string, req sdk.SetupNodeSshJSONRequestBody) error
	CloseNodeSSH(ctx context.Context, clusterID sdk.ClusterId, nodeID string) error
	GetClusterNode(ctx context.Context, clusterID sdk.ClusterId, nodeID string) (*sdk.Node, error)
}

func New(cfg *config.Config, log logrus.FieldLogger) (Interface, error) {
	accessToken := cfg.AccessToken
	apiURL := fmt.Sprintf("https://%s/v1", cfg.Hostname)

	tr := http.DefaultTransport
	if cfg.Debug {
		tr = &loggingTransport{log: log}
	}
	httpClientOption := func(client *sdk.Client) error {
		client.Client = &http.Client{
			Transport: tr,
			Timeout:   defaultTimeout,
		}
		return nil
	}
	apiTokenOption := sdk.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("X-API-Key", accessToken)
		return nil
	})
	apiClient, err := sdk.NewClientWithResponses(apiURL, httpClientOption, apiTokenOption)
	if err != nil {
		return nil, err
	}

	return &client{
		apiURL:   apiURL,
		hostname: cfg.Hostname,
		api:      apiClient,
	}, nil
}

type client struct {
	apiURL   string
	hostname string
	api      sdk.ClientWithResponsesInterface
}

func (c *client) GetClusterNode(ctx context.Context, clusterID sdk.ClusterId, nodeID string) (*sdk.Node, error) {
	resp, err := c.api.GetClusterNodeWithResponse(ctx, clusterID, nodeID)
	if err != nil {
		return nil, err
	}
	if err := c.checkResponse(resp, err, http.StatusOK); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

func (c *client) SetupNodeSSH(ctx context.Context, clusterID sdk.ClusterId, nodeID string, req sdk.SetupNodeSshJSONRequestBody) error {
	resp, err := c.api.SetupNodeSshWithResponse(ctx, clusterID, nodeID, req)
	if err != nil {
		return err
	}
	if err := c.checkResponse(resp, err, http.StatusOK); err != nil {
		return err
	}
	return nil
}

func (c *client) CloseNodeSSH(ctx context.Context, clusterID sdk.ClusterId, nodeID string) error {
	resp, err := c.api.CloseNodeSshWithResponse(ctx, clusterID, nodeID)
	if err != nil {
		return err
	}
	return c.checkResponse(resp, err, http.StatusOK)
}

func (c *client) ListClusterNodes(ctx context.Context, req sdk.ClusterId) ([]sdk.Node, error) {
	resp, err := c.api.GetClusterNodesWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := c.checkResponse(resp, err, http.StatusOK); err != nil {
		return nil, err
	}
	return resp.JSON200.Items, nil
}

func (c *client) GetCluster(ctx context.Context, req sdk.ClusterId) (*sdk.KubernetesCluster, error) {
	resp, err := c.api.GetClusterWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := c.checkResponse(resp, err, http.StatusOK); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

func (c *client) DeleteCluster(ctx context.Context, req sdk.ClusterId) error {
	resp, err := c.api.DeleteClusterWithResponse(ctx, req)
	if err != nil {
		return err
	}
	if err := c.checkResponse(resp, err, http.StatusNoContent); err != nil {
		return err
	}
	return nil
}

func (c *client) FeedbackEvents(ctx context.Context, req sdk.ClusterId) ([]sdk.KubernetesClusterFeedbackEvent, error) {
	resp, err := c.api.GetClusterFeedbackEventsWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := c.checkResponse(resp, err, http.StatusOK); err != nil {
		return nil, err
	}
	return resp.JSON200.Items, nil
}

func (c *client) CreateNewCluster(ctx context.Context, body sdk.CreateNewClusterJSONRequestBody) (*sdk.KubernetesCluster, error) {
	resp, err := c.api.CreateNewClusterWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := c.checkResponse(resp, err, http.StatusCreated); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

func (c *client) ListRegions(ctx context.Context) ([]sdk.CastRegion, error) {
	resp, err := c.api.ListRegionsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.checkResponse(resp, err, http.StatusOK); err != nil {
		return nil, err
	}
	return resp.JSON200.Items, nil
}

func (c *client) ListCloudCredentials(ctx context.Context) ([]sdk.CloudCredentials, error) {
	resp, err := c.api.ListCloudCredentialsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.checkResponse(resp, err, http.StatusOK); err != nil {
		return nil, err
	}
	return resp.JSON200.Items, nil
}

func (c *client) GetClusterKubeconfig(ctx context.Context, req sdk.ClusterId) ([]byte, error) {
	resp, err := c.api.GetClusterKubeconfigWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := c.checkResponse(resp, err, http.StatusOK); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (c *client) ListKubernetesClusters(ctx context.Context, req *sdk.ListKubernetesClustersParams) ([]sdk.KubernetesCluster, error) {
	resp, err := c.api.ListKubernetesClustersWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := c.checkResponse(resp, err, http.StatusOK); err != nil {
		return nil, err
	}
	return resp.JSON200.Items, nil
}

func (c *client) ListAuthTokens(ctx context.Context) ([]sdk.AuthToken, error) {
	resp, err := c.api.ListAuthTokensWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.checkResponse(resp, err, http.StatusOK); err != nil {
		return nil, err
	}
	return resp.JSON200.Items, nil
}

func (c *client) GetMyPublicIP(ctx context.Context) (string, error) {
	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(fmt.Sprintf("https://%s/my-public-ip", c.hostname))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func (c *client) checkResponse(resp sdk.Response, err error, expectedStatus int) error {
	if err != nil {
		return err
	}
	if resp.StatusCode() != expectedStatus {
		errBody := strings.ToLower(strings.TrimSpace(string(resp.GetBody())))
		if resp.StatusCode() == http.StatusUnauthorized {
			return fmt.Errorf("unauthorized to access %s: %s, run 'cast configure' to setup access token", c.apiURL, errBody)
		}
		if resp.StatusCode() == http.StatusInternalServerError {
			return errors.New("internal server error occurred, please try again")
		}
		return fmt.Errorf("expected status code %d, received: status=%d body=%s", expectedStatus, resp.StatusCode(), errBody)
	}
	return nil
}
