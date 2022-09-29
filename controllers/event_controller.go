package controllers

import (
	"fmt"
	"strings"

	"github.com/porter-dev/porter-agent/pkg/event"
	"github.com/porter-dev/porter-agent/pkg/incident"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// EventController listens to events from the Kubernetes API and performs the following operations concurrently:
//   1. Stores the events in the given event store
//   2. Triggers the incident detection loop
type EventController struct {
	KubeClient       *kubernetes.Clientset
	KubeVersion      incident.KubernetesVersion
	EventStore       event.EventStore
	IncidentDetector *incident.IncidentDetector
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
	if e.hasBeenProcessed(k8sEvent) {
		return nil
	}

	fmt.Println("processing kubernetes event:", k8sEvent.Name)

	filteredEvent := event.NewFilteredEventFromK8sEvent(k8sEvent)

	es := []*event.FilteredEvent{filteredEvent}

	// trigger incident detection loop
	err := e.IncidentDetector.DetectIncident(es)

	if err != nil {
		return e.updateEventCache(k8sEvent, err)
	}

	return e.updateEventCache(k8sEvent, nil)
}

func (e *EventController) hasBeenProcessed(k8sEvent *v1.Event) bool {
	// TODO: query the event cache
	return false
}

func (e *EventController) updateEventCache(k8sEvent *v1.Event, currError error) error {
	// TODO: update the event cache
	return currError
}

func (e *EventController) processDeleteEvent(obj interface{}) {
	// TODO: remove from event cache
}
