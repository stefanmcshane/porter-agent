package controllers

import (
	"fmt"
	"strings"

	"github.com/porter-dev/porter-agent/pkg/event"
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
	KubeClient *kubernetes.Clientset
	EventStore event.EventStore
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
	event := obj.(*v1.Event)
	fmt.Println("processing add event:", event.Name)
}

func (e *EventController) processUpdateEvent(oldObj, newObj interface{}) {
	event := newObj.(*v1.Event)
	fmt.Println("processing update event:", event.Name)

}

func (e *EventController) processDeleteEvent(obj interface{}) {
	event := obj.(*v1.Event)
	fmt.Println("processing delete event:", event.Name)
}
