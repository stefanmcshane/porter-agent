package lokistore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/porter-dev/porter-agent/pkg/logstore"
)

type LokiHTTPClientConf struct {
	Address string
}

type Client struct {
	client  *http.Client
	address string
}

func NewClient(conf *LokiHTTPClientConf) *Client {
	return &Client{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		address: conf.Address,
	}
}

type QueryRangeStreamResponse struct {
	Status string         `json:"status"`
	Data   QueryRangeData `json:"data"`
}

type QueryRangeData struct {
	ResultType string                 `json:"resultType"`
	Result     []QueryRangeStreamItem `json:"result"`
}

type QueryRangeStreamItem struct {
	Stream QueryRangeStreamMeta   `json:"stream"`
	Values QueryRangeStreamValues `json:"values"`
}

type QueryRangeStreamMeta struct {
	Filename string `json:"filename"`
}

type QueryRangeStreamValues [][]string

func (c *Client) QueryRange(options logstore.QueryOptions) (*QueryRangeStreamResponse, error) {
	params := make(map[string][]string)
	params["query"] = []string{
		logstore.ConstructSearch(logstore.LabelsMapToString(options.Labels, "=~", options.CustomSelectorSuffix), options.SearchParam),
	}

	params["limit"] = []string{
		fmt.Sprintf("%d", options.Limit),
	}

	params["start"] = []string{
		fmt.Sprintf("%v", options.Start.UnixNano()),
	}

	params["end"] = []string{
		fmt.Sprintf("%v", options.End.UnixNano()),
	}

	fmt.Println("start is:", options.Start.UnixNano())
	fmt.Println("end is:", options.End.UnixNano())

	resBytes, err := c.get("/loki/api/v1/query_range", params)

	if err != nil {
		return nil, err
	}

	resp := &QueryRangeStreamResponse{}

	if err := json.Unmarshal(resBytes, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) get(path string, params map[string][]string) ([]byte, error) {
	urlVals := url.Values(params)
	encodedURLVals := urlVals.Encode()

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s%s?%s", c.address, path, encodedURLVals),
		nil,
	)

	res, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	resBytes, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	fmt.Println("loki response bytes are:", string(resBytes))

	return resBytes, nil
}
