package job

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/porter-dev/porter-agent/internal/logger"
	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/internal/repository"
	"github.com/porter-dev/porter-agent/pkg/event"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type JobEventProducer struct {
	KubeClient kubernetes.Clientset
	Repository *repository.Repository
	Logger     *logger.Logger
}

func (j *JobEventProducer) ParseFilteredEvents(es []*event.FilteredEvent) error {
	for _, e := range es {
		// this parser only looks for low-severity events
		if e.Severity != event.EventSeverityLow {
			continue
		}

		// we only look at a subset of kubernetes reasons
		if r := e.KubernetesReason; r != "Completed" && r != "Running" {
			continue
		}

		// de-duplicate the event
		if j.isInCache(e) {
			continue
		}

		// place the event in the cache as soon as possible
		now := time.Now()

		j.Repository.JobCache.CreateJobCache(&models.JobCache{
			PodName:      e.PodName,
			PodNamespace: e.PodNamespace,
			Reason:       e.KubernetesReason,
			Timestamp:    &now,
		})

		// populate the event
		err := e.Populate(j.KubeClient)

		if err != nil {
			return err
		}

		// only parse events with a job owner - this is checked in the events but we check this here as well
		if strings.ToLower(e.Owner.Kind) == "job" {
			// case on the reason and store the events
			var porterEvent *models.Event

			switch e.KubernetesReason {
			case "Running":
				porterEvent = models.NewJobStartedEventV1()
			case "Completed":
				porterEvent = models.NewJobFinishedEventV1()
			}

			porterEvent.ReleaseName = e.ReleaseName
			porterEvent.ReleaseNamespace = e.PodNamespace
			porterEvent.Timestamp = e.Timestamp
			porterEvent.AdditionalQueryMeta = fmt.Sprintf("job/%s", e.Owner.Name)

			jobEventData := podToJobEventData(e.Pod)

			eventDataBytes, err := json.Marshal(jobEventData)

			if err != nil {
				return err
			}

			porterEvent.Data = eventDataBytes

			// check cache hits again in case this has been added since checking it above
			if j.isInCache(e) {
				continue
			}

			porterEvent, err = j.Repository.Event.CreateEvent(porterEvent)

			if err != nil {
				j.Logger.Error().Caller().Msgf("could not save new event: %s", err.Error())
				return err
			}
		}
	}

	return nil

}

func (j *JobEventProducer) isInCache(e *event.FilteredEvent) bool {
	if cacheHits, err := j.Repository.JobCache.ListJobCaches(e.PodName, e.PodNamespace, e.KubernetesReason); err == nil && len(cacheHits) > 0 {
		return true
	}

	return false
}

// We strip out the spec of the pod which could include sensitive information, retaining only
// the metadata and the status
type JobEventData struct {
	Meta   *metav1.ObjectMeta `json:"metadata"`
	Status *v1.PodStatus      `json:"status"`
}

func podToJobEventData(pod *v1.Pod) *JobEventData {
	return &JobEventData{
		Meta:   &pod.ObjectMeta,
		Status: &pod.Status,
	}
}
