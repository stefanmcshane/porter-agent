package types

import "github.com/porter-dev/porter-agent/pkg/logstore/lokistore"

type GetStatusResponse struct {
	Loki lokistore.Reachability `json:"loki"`
}
