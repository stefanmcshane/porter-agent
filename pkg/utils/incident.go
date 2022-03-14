package utils

import (
	"fmt"
	"strings"
)

// incidents are of the form "incident:<pod_name>:<release_name>:<namespace>"
type Incident struct {
	releaseName string
	namespace   string
}

func NewIncident(releaseName, namespace string) *Incident {
	return &Incident{
		releaseName: releaseName,
		namespace:   namespace,
	}
}

func NewIncidentFromString(id string) (*Incident, error) {
	segments := strings.Split(id, ":")

	if len(segments) != 3 || (len(segments) > 0 && segments[0] != "incident") {
		return nil, fmt.Errorf("invalid incident of the form: %s", id)
	}

	return &Incident{
		releaseName: segments[1],
		namespace:   segments[2],
	}, nil
}

func (inc *Incident) GetReleaseName() string {
	return inc.releaseName
}

func (inc *Incident) GetNamespace() string {
	return inc.namespace
}

func (inc *Incident) ToString() string {
	return fmt.Sprintf("incident:%s:%s", inc.GetReleaseName(), inc.GetNamespace())
}
