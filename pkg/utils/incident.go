package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// incidents are of the form "incident:<release_name>:<namespace>:<timestamp>"
type Incident struct {
	releaseName string
	namespace   string
	timestamp   int64
}

func NewIncident(releaseName, namespace string, timestamp int64) *Incident {
	return &Incident{
		releaseName: releaseName,
		namespace:   namespace,
		timestamp:   timestamp,
	}
}

func NewIncidentFromString(id string) (*Incident, error) {
	segments := strings.Split(id, ":")

	if len(segments) != 4 || (len(segments) > 0 && segments[0] != "incident") {
		return nil, fmt.Errorf("invalid incident of the form: %s", id)
	}

	incident := &Incident{
		releaseName: segments[1],
		namespace:   segments[2],
	}

	timestamp, err := strconv.ParseInt(segments[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error converting last segment to int64 of incident ID: %s", id)
	}

	incident.timestamp = timestamp

	return incident, nil
}

func (inc *Incident) GetReleaseName() string {
	return inc.releaseName
}

func (inc *Incident) GetNamespace() string {
	return inc.namespace
}

func (inc *Incident) GetTimestamp() int64 {
	return inc.timestamp
}

func (inc *Incident) GetTimestampAsTime() time.Time {
	return time.Unix(inc.timestamp, 0)
}

func (inc *Incident) ToString() string {
	return fmt.Sprintf("incident:%s:%s:%d", inc.GetReleaseName(), inc.GetNamespace(), inc.GetTimestamp())
}
