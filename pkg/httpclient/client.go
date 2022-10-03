package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/internal/logger"
)

type HTTPClientConf struct {
	PorterHost  string `env:"PORTER_HOST"`
	PorterToken string `env:"PORTER_TOKEN"`
	ClusterID   string `env:"CLUSTER_ID"`
	ProjectID   string `env:"PROJECT_ID"`
}

type Client struct {
	client               *http.Client
	token                string
	host                 string
	projectID, clusterID string
	logger               *logger.Logger
}

func NewClient(conf *HTTPClientConf, logger *logger.Logger) *Client {
	return &Client{
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
		token:     conf.PorterToken,
		host:      conf.PorterHost,
		projectID: conf.ProjectID,
		clusterID: conf.ClusterID,
		logger:    logger,
	}
}

func (c *Client) NotifyNew(incident *types.Incident) error {
	_, err := c.post(fmt.Sprintf("/api/projects/%s/clusters/%s/incidents/notify_new", c.projectID, c.clusterID), incident)

	return err
}

func (c *Client) NotifyResolved(incident *types.Incident) error {
	_, err := c.post(fmt.Sprintf("/api/projects/%s/clusters/%s/incidents/notify_resolved", c.projectID, c.clusterID), incident)

	return err
}

func (c *Client) post(path string, body interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s%s", c.host, path)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/json")

	return c.client.Do(req)
}
