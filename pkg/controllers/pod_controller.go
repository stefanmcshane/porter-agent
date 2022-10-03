package controllers

import (
	"strings"

	"github.com/porter-dev/porter-agent/internal/logger"
	"github.com/porter-dev/porter-agent/pkg/event"
	"github.com/porter-dev/porter-agent/pkg/incident"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// PodController watches pods on the cluster and generates events for those pods. It triggers the incident detection loop
// when pods are updated on the cluster.
type PodController struct {
	KubeClient       *kubernetes.Clientset
	EventStore       event.EventStore
	KubeVersion      incident.KubernetesVersion
	IncidentDetector *incident.IncidentDetector
	Logger           *logger.Logger
}

func (p *PodController) Start() {
	tweakListOptionsFunc := func(options *metav1.ListOptions) {}

	factory := informers.NewSharedInformerFactoryWithOptions(
		p.KubeClient,
		0,
		informers.WithTweakListOptions(tweakListOptionsFunc),
	)

	informer := factory.Core().V1().Pods().Informer()

	stopper := make(chan struct{})
	errorchan := make(chan error)

	informer.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
		if strings.HasSuffix(err.Error(), ": Unauthorized") {
			errorchan <- &AuthError{}
		}
	})

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: p.processUpdatePod,
		AddFunc:    p.processAddPod,
		DeleteFunc: p.processDeletePod,
	})

	p.Logger.Info().Caller().Msgf("started pod controller")

	informer.Run(stopper)
}

func (p *PodController) processAddPod(obj interface{}) {
	pod := obj.(*v1.Pod)
	p.processPod(pod)
}

func (p *PodController) processUpdatePod(oldObj, newObj interface{}) {
	pod := newObj.(*v1.Pod)
	p.processPod(pod)
}

func (p *PodController) processPod(pod *v1.Pod) error {
	p.Logger.Info().Caller().Msgf("processing pod: %s", pod.Name)

	es := event.NewFilteredEventsFromPod(pod)

	// trigger incident detection loop
	err := p.IncidentDetector.DetectIncident(es)

	if err != nil {
		return err
	}

	return nil
}

func (p *PodController) processDeletePod(obj interface{}) {
	// do nothing
}
