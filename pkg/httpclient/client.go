package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/spf13/viper"
)

var (
	porterHost  string
	porterPort  string
	porterToken string
	clusterID   string
	projectID   string
)

func init() {
	viper.SetDefault("PORTER_PORT", "80")
	viper.AutomaticEnv()

	porterPort = viper.GetString("PORTER_PORT")
	porterHost = getStringOrDie("PORTER_HOST")
	porterToken = getStringOrDie("PORTER_TOKEN")
	clusterID = getStringOrDie("CLUSTER_ID")
	projectID = getStringOrDie("PROJECT_ID")

}

func getStringOrDie(key string) string {
	value := viper.GetString(key)

	if value == "" {
		panic(fmt.Errorf("empty %s", key))
		// consumerLog.Error(fmt.Errorf("empty %s", key), fmt.Sprintf("%s must not be empty", key))
		// os.Exit(1)
	}

	return value
}

type ClientOptions struct{}

type Client struct {
	client               *http.Client
	token                string
	host                 string
	projectID, clusterID string
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
		token:     porterToken,
		host:      porterHost,
		projectID: projectID,
		clusterID: clusterID,
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

func (c *Client) get(url string, options ...ClientOptions) (*http.Response, error) {
	return c.client.Get(fmt.Sprintf("%s%s", c.host, url))
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
