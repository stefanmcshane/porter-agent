package event

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

type EventSeverity string

const (
	EventSeverityCritical EventSeverity = "critical"
	EventSeverityHigh     EventSeverity = "high"
	EventSeverityLow      EventSeverity = "low"
)

type EventSource string

const (
	Pod      EventSource = "pod"
	K8sEvent EventSource = "event"
)

type FilteredEvent struct {
	Source EventSource

	PodName      string
	PodNamespace string

	KubernetesReason  string
	KubernetesMessage string

	Severity EventSeverity

	Timestamp *time.Time

	// (optional) The exit code of the application, if applicable
	ExitCode uint

	// (optional) The pod config, if applicable or present
	// TODO
	Pod *v1.Pod

	// (optional) The owner data, if applicable or present
	Owner *EventOwner
}

type EventOwner struct {
	Namespace, Name, Kind string
}

// SetPodData is used to set the data for the pod directly. This is useful for cases where querying the
// live status of the pod via PopulatePodData may fail if the pod has been deleted.
func (e *FilteredEvent) SetPodData(pod *v1.Pod) {
	e.Pod = pod
}

func (e *FilteredEvent) PopulatePodData(k8sClient kubernetes.Clientset) error {
	if e.Pod != nil {
		return nil
	}

	pod, err := k8sClient.CoreV1().Pods(e.PodNamespace).Get(
		context.Background(),
		e.PodName,
		metav1.GetOptions{},
	)

	if err != nil {
		return err
	}

	e.Pod = pod
	return nil
}

func (e *FilteredEvent) PopulateEventOwner(k8sClient kubernetes.Clientset) error {
	// determine if pod is owned by a ReplicaSet or Job
	if e.Pod == nil {
		err := e.PopulatePodData(k8sClient)

		if err != nil {
			return err
		}
	}

	if len(e.Pod.OwnerReferences) != 1 {
		return fmt.Errorf("unable to populate event owner: pod has multiple owners")
	}

	switch o := e.Pod.OwnerReferences[0]; strings.ToLower(o.Kind) {
	case "replicaset":
		rs, err := k8sClient.AppsV1().ReplicaSets(e.PodNamespace).Get(
			context.Background(),
			o.Name,
			metav1.GetOptions{},
		)

		if err != nil {
			return err
		}

		if len(rs.OwnerReferences) != 1 {
			return fmt.Errorf("unable to populate event owner: replicaset has multiple owners")
		}

		if strings.ToLower(rs.OwnerReferences[0].Kind) != "deployment" {
			return fmt.Errorf("only replicasets with deployment owners are supported")
		}

		e.Owner = &EventOwner{
			Namespace: e.PodNamespace,
			Name:      rs.OwnerReferences[0].Name,
			Kind:      rs.OwnerReferences[0].Kind,
		}

		return nil
	case "job":
		e.Owner = &EventOwner{
			Namespace: e.PodNamespace,
			Name:      o.Name,
			Kind:      o.Kind,
		}
	}

	return fmt.Errorf("unsupported owner reference kind")
}

type EventStore interface {
	Store(e *FilteredEvent) error
	GetEventsByPodName(namespace, name string) *FilteredEvent
	GetEventsByOwner(owner *EventOwner) *FilteredEvent
}

func NewFilteredEventFromK8sEvent(k8sEvent *v1.Event) *FilteredEvent {
	var severity EventSeverity

	if k8sEvent.Type == "Normal" {
		severity = EventSeverityLow
	} else if k8sEvent.Type == "Warning" {
		severity = EventSeverityHigh
	}

	return &FilteredEvent{
		Source:            K8sEvent,
		PodName:           k8sEvent.InvolvedObject.Name,
		PodNamespace:      k8sEvent.InvolvedObject.Namespace,
		KubernetesReason:  k8sEvent.Reason,
		KubernetesMessage: k8sEvent.Message,
		Severity:          severity,
		Timestamp:         &k8sEvent.LastTimestamp.Time,
	}
}

func NewFilteredEventsFromPod(pod *v1.Pod) []*FilteredEvent {
	res := make([]*FilteredEvent, 0)

	// if the pod has failed to get scheduled in over 15 minutes, we generate a high-severity event
	for _, condition := range pod.Status.Conditions {
		if condition.Type == "PodScheduled" && (condition.Status == v1.ConditionFalse || condition.Status == v1.ConditionUnknown) {
			now := time.Now()

			// check if the last transition time was 15 minutes ago
			if condition.LastTransitionTime.Time.After(now.Add(-15 * time.Minute)) {
				res = append(res, &FilteredEvent{
					Source:            Pod,
					PodName:           pod.Name,
					PodNamespace:      pod.Namespace,
					KubernetesReason:  "Pending",
					KubernetesMessage: fmt.Sprintf("Pod has been pending for %f minutes", now.Sub(condition.LastTransitionTime.Time).Minutes()),
					Severity:          EventSeverityHigh,
					Timestamp:         &now,
				})
			}
		}
	}

	// if one or more containers failed to start, we generate a set of events
	for _, containerStatus := range pod.Status.ContainerStatuses {
		// if the container is currently in a waiting state, we check to see if the last state is terminated -
		// if so, we generate an event
		if waitingState := containerStatus.State.Waiting; waitingState != nil {
			// if the waiting state is an image error, we store this as an event as well
			if waitingState.Reason == "ImagePullBackOff" || waitingState.Reason == "ErrImagePull" || waitingState.Reason == "InvalidImageName" {
				res = append(res, &FilteredEvent{
					Source:           Pod,
					PodName:          pod.Name,
					PodNamespace:     pod.Namespace,
					KubernetesReason: waitingState.Reason,
					Severity:         EventSeverityHigh,
					// We set this to the creation timestamp of the pod - note that this will miss cases where the image has been
					// deleted from the registry and the pod was restarted afterwards.
					Timestamp: &pod.CreationTimestamp.Time,
				})
			}

			if lastTermState := containerStatus.LastTerminationState.Terminated; lastTermState != nil {
				// add the last termination state as an event
				res = append(res, getEventFromTerminationState(pod.Name, pod.Namespace, lastTermState))
			}
		} else if termState := containerStatus.State.Terminated; termState != nil {
			res = append(res, getEventFromTerminationState(pod.Name, pod.Namespace, termState))
		}
	}

	return res
}

func getEventFromTerminationState(podName, podNamespace string, termState *v1.ContainerStateTerminated) *FilteredEvent {
	event := &FilteredEvent{
		Source:       Pod,
		PodName:      podName,
		PodNamespace: podNamespace,
		Severity:     EventSeverityHigh,
		Timestamp:    &termState.FinishedAt.Time,
		ExitCode:     uint(termState.ExitCode),
	}

	if termState.Reason == "" {
		if termState.ExitCode != 0 {
			event.KubernetesReason = "ApplicationError"
			return event
		}
	} else {
		event.KubernetesReason = termState.Reason
		return event
	}

	return nil
}
