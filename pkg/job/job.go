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

			eventData, err := json.Marshal(e.Pod)

			if err != nil {
				return err
			}

			porterEvent.Data = eventData

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

		// add the event to the cache
		now := time.Now()

		j.Repository.JobCache.CreateJobCache(&models.JobCache{
			PodName:      e.PodName,
			PodNamespace: e.PodNamespace,
			Reason:       e.KubernetesReason,
			Timestamp:    &now,
		})
	}

	return nil

}

func (j *JobEventProducer) isInCache(e *event.FilteredEvent) bool {
	if cacheHits, err := j.Repository.JobCache.ListJobCaches(e.PodName, e.PodNamespace, e.KubernetesReason); err == nil && len(cacheHits) > 0 {
		return true
	}

	return false
}
