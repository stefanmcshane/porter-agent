package models

import (
	"encoding/json"
	"time"

	"github.com/porter-dev/porter-agent/api/server/types"
	"gorm.io/gorm"
)

type Event struct {
	gorm.Model

	UniqueID string `gorm:"unique"`

	Version          string
	Type             types.EventType
	ReleaseName      string
	ReleaseNamespace string
	Timestamp        *time.Time

	Data []byte
}

func NewIncidentEventV1() *Event {
	randStr, _ := GenerateRandomBytes(16)

	return &Event{
		UniqueID: randStr,
		Type:     types.EventTypeIncident,
		Version:  "v1",
	}
}

func NewDeploymentFinishedEventV1() *Event {
	randStr, _ := GenerateRandomBytes(16)

	return &Event{
		UniqueID: randStr,
		Type:     types.EventTypeDeploymentFinished,
		Version:  "v1",
	}
}

func (e *Event) ToAPIType() *types.Event {
	eventData := make(map[string]interface{})

	json.Unmarshal(e.Data, &eventData)

	return &types.Event{
		Version:          e.Version,
		Type:             e.Type,
		ReleaseName:      e.ReleaseName,
		ReleaseNamespace: e.ReleaseNamespace,
		Timestamp:        e.Timestamp,
		Data:             eventData,
	}
}
