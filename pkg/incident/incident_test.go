package incident_test

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/porter-dev/porter-agent/pkg/event"
	"github.com/porter-dev/porter-agent/pkg/incident"
	v1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
)

// 1.20 sample events

// Sample events for pod crashing due to invalid start command

// Sample events for pod crashing due to an application-level error

// Sample events for pod crashing due to OOM

// Sample events for pod failing to start due to being stuck in pending

// Sample events for pod failing to start due to image pull errors

// Ensure that pods failing their liveness probes are matched by alerting
func TestEventFailingLivenessProbeMatch1_20(t *testing.T) {
	events := loadEventFixtures(t, "failing-liveness-probe.json")

	matches := make([]*incident.EventMatch, 0)

	for _, event := range events {
		currMatch := incident.GetEventMatchFromEvent(incident.KubernetesVersion_1_20, event)

		if currMatch != nil {
			matches = append(matches, currMatch)
		}
	}

	// assert that only a single match was found
	assert.Equal(t, 1, len(matches), "only a single match should exist")

	// match should be liveness probe
	assert.Equal(t, "The application is failing its health check", matches[0].Summary, "match should be liveness probe failure")
}

// Ensure that pods stuck in a generic crash loop are matched by alerting
func TestEventBackOff1_20(t *testing.T) {
	events := loadEventFixtures(t, "generic-crash-loop.json")

	matches := get1_20Matches(t, events)

	// assert that only a single match was found
	assert.Equal(t, 1, len(matches), "only a single match should exist")

	// match should be generic back off error
	assert.Equal(t, "The application was restarted", matches[0].Summary, "match should be generic back off error")
}

// Ensure that pods which are killed due to OOM are matched by alerting
func TestPodOOMKilled1_20(t *testing.T) {
	events := loadPodFixtures(t, "oom-killed.json")

	matches := get1_20Matches(t, events)

	// assert that only a single match was found
	assert.Equal(t, 1, len(matches), "only a single match should exist")

	// match should be oom killed
	assert.Equal(t, "The application ran out of memory", matches[0].Summary, "match should be OOM")
}

// Ensure that pods killed due to an app error have events stored
func TestPodAppError1_20(t *testing.T) {
	events := loadPodFixtures(t, "non-zero-exit-code.json")

	matches := get1_20Matches(t, events)

	// assert that only a single match was found
	assert.Equal(t, 1, len(matches), "only a single match should exist")

	// match should be app error
	assert.Equal(t, "The application exited with a non-zero exit code", matches[0].Summary, "match should be app error")
}

// Ensure that pods which are killed due to both app error and then face an image pull error have both events stored
func TestPodImagePullBackOffAndAppError1_20(t *testing.T) {
	events := loadPodFixtures(t, "image-pull-backoff-after-app-error.json")

	matches := get1_20Matches(t, events)

	// assert that two events have matched
	assert.Equal(t, 2, len(matches), "two matches should exist")

	// one of the matches should be an application error, the other match should be an image error
	numProcessed := 0

	for _, match := range matches {
		if match.Summary == "The application has an invalid image" || match.Summary == "The application exited with a non-zero exit code" {
			numProcessed++
		}
	}

	assert.Equal(t, 2, numProcessed, "both application error and image pull backoff error should be matched")
}

func get1_20Matches(t *testing.T, events []*event.FilteredEvent) []*incident.EventMatch {
	matches := make([]*incident.EventMatch, 0)

	for _, event := range events {
		currMatch := incident.GetEventMatchFromEvent(incident.KubernetesVersion_1_20, event)

		if currMatch != nil {
			matches = append(matches, currMatch)
		}
	}

	return matches
}

func loadEventFixtures(t *testing.T, filename string) []*event.FilteredEvent {
	_, currentFileName, _, _ := runtime.Caller(0)

	fullPath := filepath.Join(filepath.Dir(currentFileName), "fixtures/1.20/events", filename)

	fileBytes, err := ioutil.ReadFile(fullPath)

	if err != nil {
		t.Fatalf("Could not read file: %v", err)
	}

	k8sEvents := make([]*v1.Event, 0)

	err = json.Unmarshal(fileBytes, &k8sEvents)

	if err != nil {
		t.Fatalf("Could not convert file to JSON: %v", err)
	}

	filteredEvents := make([]*event.FilteredEvent, 0)

	for _, k8sEvent := range k8sEvents {
		filteredEvent := event.NewFilteredEventFromK8sEvent(k8sEvent)

		if filteredEvent != nil {
			filteredEvents = append(filteredEvents, filteredEvent)
		}
	}

	return filteredEvents
}

func loadPodFixtures(t *testing.T, filename string) []*event.FilteredEvent {
	_, currentFileName, _, _ := runtime.Caller(0)

	fullPath := filepath.Join(filepath.Dir(currentFileName), "fixtures/1.20/pods", filename)

	fileBytes, err := ioutil.ReadFile(fullPath)

	if err != nil {
		t.Fatalf("Could not read file: %v", err)
	}

	pod := &v1.Pod{}

	err = json.Unmarshal(fileBytes, pod)

	if err != nil {
		t.Fatalf("Could not convert file to JSON: %v", err)
	}

	return event.NewFilteredEventsFromPod(pod)
}

// Sample events for pods restarting due to failing startup probe

// Sample events for pods not ready due to failing readiness probe
