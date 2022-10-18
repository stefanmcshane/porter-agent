package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/porter-dev/porter-agent/internal/logger"
	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/internal/repository"
	"github.com/porter-dev/porter-agent/pkg/event"
	"github.com/porter-dev/porter-agent/pkg/incident"
	"github.com/porter-dev/porter-agent/pkg/logstore"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// EventController listens to events from the Kubernetes API and performs the following operations concurrently:
//  1. Stores the events in the given event store
//  2. Triggers the incident detection loop
type EventController struct {
	KubeClient       *kubernetes.Clientset
	KubeVersion      incident.KubernetesVersion
	EventStore       event.EventStore
	IncidentDetector *incident.IncidentDetector
	Repository       *repository.Repository
	LogStore         logstore.LogStore
	Logger           *logger.Logger
}

type AuthError struct{}

func (e *AuthError) Error() string {
	return "Unauthorized error"
}

func (e *EventController) Start() {
	tweakListOptionsFunc := func(options *metav1.ListOptions) {
		options.FieldSelector = "involvedObject.kind=Pod"
	}

	factory := informers.NewSharedInformerFactoryWithOptions(
		e.KubeClient,
		0,
		informers.WithTweakListOptions(tweakListOptionsFunc),
	)

	informer := factory.Core().V1().Events().Informer()

	stopper := make(chan struct{})
	errorchan := make(chan error)

	informer.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
		if strings.HasSuffix(err.Error(), ": Unauthorized") {
			errorchan <- &AuthError{}
		}
	})

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: e.processUpdateEvent,
		AddFunc:    e.processAddEvent,
		DeleteFunc: e.processDeleteEvent,
	})

	e.Logger.Info().Caller().Msgf("started event controller")

	informer.Run(stopper)
}

func (e *EventController) processAddEvent(obj interface{}) {
	k8sEvent := obj.(*v1.Event)
	e.processEvent(k8sEvent)
}

func (e *EventController) processUpdateEvent(oldObj, newObj interface{}) {
	k8sEvent := newObj.(*v1.Event)
	e.processEvent(k8sEvent)
}

func (e *EventController) processEvent(k8sEvent *v1.Event) error {
	// TODO: de-duplicate events which have already been stored/processed based
	// on both the timestamp and the event ID
	hasBeenProcessed, revision := e.hasBeenProcessed(k8sEvent)

	e.Logger.Info().Caller().Msgf("processing kubernetes event: %s", k8sEvent.Name)

	// if we do not have a revision, we must query for the revision
	if revision == "" {
		pod, err := e.KubeClient.CoreV1().Pods(k8sEvent.InvolvedObject.Namespace).Get(
			context.Background(),
			k8sEvent.InvolvedObject.Name,
			metav1.GetOptions{},
		)

		if err != nil {
			e.Logger.Error().Caller().Msgf("could not query for pod event: %v", err)
			return e.updateEventCache(k8sEvent, revision, err)
		}

		revision = pod.Annotations["helm.sh/revision"]
	}

	// store the event via the log store
	if serializedEvent, err := serializeEvent(k8sEvent); err == nil {
		err = e.LogStore.Push(map[string]string{
			"event_store":      "true",
			"pod":              k8sEvent.InvolvedObject.Name,
			"namespace":        k8sEvent.InvolvedObject.Namespace,
			"helm_sh_revision": revision,
			"porter_pod_name":  fmt.Sprintf("%s_%s", k8sEvent.InvolvedObject.Name, revision),
		}, serializedEvent, time.Now())

		if err != nil {
			e.Logger.Error().Caller().Msgf("could not push serialized event to log store: %v", err)
			return e.updateEventCache(k8sEvent, revision, err)
		}
	}

	// skip the filtered event detection if the event has been processed
	if !hasBeenProcessed {
		filteredEvent := event.NewFilteredEventFromK8sEvent(k8sEvent)

		es := []*event.FilteredEvent{filteredEvent}

		// trigger incident detection loop
		err := e.IncidentDetector.DetectIncident(es)

		if err != nil {
			e.Logger.Error().Caller().Msgf("encountered error while running incident detection: %v", err)
			return e.updateEventCache(k8sEvent, revision, err)
		}
	}

	return e.updateEventCache(k8sEvent, revision, nil)
}

func (e *EventController) hasBeenProcessed(k8sEvent *v1.Event) (bool, string) {
	caches, _ := e.Repository.EventCache.ListEventCachesForEvent(getEventCacheID(k8sEvent))

	if len(caches) > 0 {
		return true, caches[0].Revision
	}

	return false, ""
}

func (e *EventController) updateEventCache(k8sEvent *v1.Event, revision string, currError error) error {
	now := time.Now()

	_, err := e.Repository.EventCache.CreateEventCache(&models.EventCache{
		EventUID:     getEventCacheID(k8sEvent),
		PodName:      k8sEvent.InvolvedObject.Name,
		PodNamespace: k8sEvent.InvolvedObject.Namespace,
		Timestamp:    &now,
		Revision:     revision,
	})

	if err != nil {
		e.Logger.Error().Caller().Msgf("could not create event in event cache: %v", err)
	}

	return currError
}

func (e *EventController) processDeleteEvent(obj interface{}) {
	// do nothing
}

func getEventCacheID(k8sEvent *v1.Event) string {
	return fmt.Sprintf("%v-%s-%s-%s", k8sEvent.UID, k8sEvent.Name, k8sEvent.Namespace, k8sEvent.InvolvedObject.Name)
}

func serializeEvent(k8sEvent *v1.Event) (string, error) {
	// set the managed fields to null, as this adds a lot of unnecessary data to serialized
	// object
	k8sEvent.ObjectMeta.ManagedFields = nil

	jsonBytes, err := json.Marshal(k8sEvent)

	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
