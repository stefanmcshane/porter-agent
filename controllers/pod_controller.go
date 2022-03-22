/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/porter-dev/porter-agent/pkg/models"
	"github.com/porter-dev/porter-agent/pkg/processor"
	"github.com/porter-dev/porter-agent/pkg/redis"
	"github.com/porter-dev/porter-agent/pkg/utils"
	"github.com/spf13/viper"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	redisHost        string
	redisPort        string
	maxTailLines     int64
	containerSignals map[int32]string
)

func init() {
	viper.SetDefault("REDIS_HOST", "porter-redis-master")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("MAX_TAIL_LINES", int64(100))
	viper.AutomaticEnv()

	redisHost = viper.GetString("REDIS_HOST")
	redisPort = viper.GetString("REDIS_PORT")
	maxTailLines = viper.GetInt64("MAX_TAIL_LINES")

	// refer: https://www.man7.org/linux/man-pages/man7/signal.7.html
	containerSignals = make(map[int32]string)
	containerSignals[1] = "SIGHUP"
	containerSignals[2] = "SIGINT"
	containerSignals[3] = "SIGQUIT"
	containerSignals[4] = "SIGILL"
	containerSignals[5] = "SIGTRAP"
	containerSignals[6] = "SIGABRT"
	containerSignals[9] = "SIGKILL"
	containerSignals[11] = "SIGSEGV"
	containerSignals[15] = "SIGTERM"
}

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Processor processor.Interface

	redisClient *redis.Client
	KubeClient  *kubernetes.Clientset

	logger logr.Logger
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = log.FromContext(ctx)

	if r.redisClient == nil {
		r.redisClient = redis.NewClient(redisHost, redisPort, "", "", redis.PODSTORE, maxTailLines)
	}

	instance := &corev1.Pod{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if instance.Namespace == "cert-manager" || instance.Namespace == "ingress-nginx" ||
		instance.Namespace == "kube-node-lease" || instance.Namespace == "kube-public" ||
		instance.Namespace == "kube-system" || instance.Namespace == "monitoring" ||
		instance.Namespace == "porter-agent-system" {
		return ctrl.Result{}, nil
	}

	agentCreationTimestamp, err := r.redisClient.GetAgentCreationTimestamp(ctx)
	if err != nil {
		r.logger.Error(err, "redisClient.GetAgentCreationTimestamp ERROR")
		return ctrl.Result{}, err
	}

	if instance.GetCreationTimestamp().Unix() < agentCreationTimestamp {
		return ctrl.Result{}, nil
	}

	ownerName, ownerKind := r.getOwnerDetails(ctx, req, instance)

	customFinalizer := "porter.run/agent-finalizer"

	finalizers := instance.Finalizers
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		found := false
		for _, fin := range finalizers {
			if fin == customFinalizer {
				found = true
				break
			}
		}

		if !found {
			instance.SetFinalizers(append(finalizers, customFinalizer))
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{Requeue: true}, fmt.Errorf("error adding custom finalizer: %w", err)
			}
		}
	} else {
		found := false
		for _, fin := range finalizers {
			if fin == customFinalizer {
				found = true
				break
			}
		}

		if found {
			incidentID, err := r.redisClient.GetActiveIncident(ctx, ownerName, instance.Namespace)
			if err == nil {
				r.redisClient.SetPodResolved(ctx, instance.Name, incidentID) // FIXME: make use of the error
			}
		}

		return ctrl.Result{}, nil
	}

	// check latest condition by sorting
	// r.logger.Info("pod conditions before sorting", "conditions", instance.Status.Conditions)
	utils.PodConditionsSorter(instance.Status.Conditions, true)
	// r.logger.Info("pod conditions after sorting", "conditions", instance.Status.Conditions)

	// in case status conditions are not yet set for the pod
	// reconcile the event
	if len(instance.Status.Conditions) == 0 {
		r.logger.Info("empty status conditions....reconciling")
		return ctrl.Result{Requeue: true}, nil
	}

	// latestCondition := instance.Status.Conditions[0]

	if len(instance.Status.ContainerStatuses) == 0 {
		r.logger.Info("nothing in container statuses, reconciling")
		return ctrl.Result{Requeue: true}, nil
	}

	// else we still have the object
	// we can log the current condition

	reason := string(instance.Status.Phase)
	if instance.Status.Reason != "" {
		reason = instance.Status.Reason
	}

	r.logger.Info("creating container events")
	containerEvents := make(map[string]*models.ContainerEvent)

	for i := len(instance.Status.ContainerStatuses) - 1; i >= 0; i-- {
		status := instance.Status.ContainerStatuses[i]
		containerReason := ""

		if status.State.Waiting != nil && status.State.Waiting.Reason != "" {
			reason := getFilteredReason(status.State.Waiting.Reason)
			containerReason = reason
		} else if status.State.Terminated != nil && status.State.Terminated.Reason != "" {
			reason = getFilteredReason(status.State.Terminated.Reason)
			containerReason = reason
		} else if status.State.Terminated != nil && status.State.Terminated.Reason == "" {
			if status.State.Terminated.Signal != 0 {
				reason = fmt.Sprintf("Non-zero signal: %d", status.State.Terminated.Signal)
			} else {
				reason = fmt.Sprintf("Non-zero exit code: %d", status.State.Terminated.ExitCode)
			}
		}

		if status.LastTerminationState.Terminated != nil {
			event := &models.ContainerEvent{
				Name:     status.Name,
				ExitCode: status.LastTerminationState.Terminated.ExitCode,
			}

			if containerReason != "" {
				event.Reason = containerReason
			}

			if event.Reason == "" {
				if signal, ok := containerSignals[event.ExitCode]; ok {
					event.Reason = fmt.Sprintf("Container exited with %s signal. This is most probably a system error.", signal)
				} else {
					event.Reason = fmt.Sprintf("Container exited with %d exit code", event.ExitCode)

					if event.ExitCode > 128 {
						event.Reason = fmt.Sprintf("%s. This is most probably a system error.", event.Reason)
					} else {
						event.Reason = fmt.Sprintf("%s. This is most probably an application error.", event.Reason)
					}
				}
			}

			event.Message = status.LastTerminationState.Terminated.Message

			containerEvents[status.Name] = event
		}
	}

	if len(containerEvents) == 0 {
		// check if pod's container is in waiting state, might be another container error
		for i := len(instance.Status.ContainerStatuses) - 1; i >= 0; i-- {
			status := instance.Status.ContainerStatuses[i]

			if status.State.Waiting != nil && status.State.Waiting.Reason != "" {
				reason = getFilteredReason(status.State.Waiting.Reason)

				if !strings.HasPrefix(reason, "Kubernetes error:") {
					// only consider it as an error if it is in our monitoring list of reasons
					containerEvents[status.Name] = &models.ContainerEvent{
						Name:    status.Name,
						Reason:  reason,
						Message: status.State.Waiting.Message,
						LogID:   "NOOP",
					}
				}
			}
		}
	}

	if len(containerEvents) == 0 {
		incidentID, err := r.redisClient.GetActiveIncident(ctx, ownerName, instance.Namespace)
		if err == nil {
			r.redisClient.SetPodResolved(ctx, instance.Name, incidentID) // FIXME: make use of the error

			if ownerKind == "Job" {
				// since a job has one running pod at a time and here we know that it has run successfully
				r.redisClient.SetJobIncidentResolved(ctx, incidentID) // FIXME: make use of the error
			}
		}

		return ctrl.Result{}, nil // FIXME: better introspection to requeue here
	}

	r.logger.Info("container events created.", "length", len(containerEvents))

	incidentID, isNewIncident, err := r.redisClient.GetOrCreateActiveIncident(ctx, ownerName, instance.Namespace)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	r.logger.Info("active incident ID", "incidentID", incidentID)

	event := &models.PodEvent{
		PodName:         instance.Name,
		Namespace:       instance.Namespace,
		OwnerName:       ownerName,
		OwnerType:       ownerKind,
		Timestamp:       time.Now().Unix(),
		Phase:           string(instance.Status.Phase),
		ContainerEvents: containerEvents,
		Reason:          reason,
	}

	event.Status = fmt.Sprintf("Type: %s, Status: %s", instance.Status.Conditions[0].Type,
		instance.Status.Conditions[0].Status)
	_, event.Message = r.getReasonAndMessage(instance, event.OwnerType)

	// if event.Reason == "" {
	// 	// FIXME: do better
	// 	event.Reason = "The pod terminated due to an error"
	// }

	// if event.Message == "" {
	// 	// FIXME: do better
	// 	event.Message = "We were unable to find the exact issue with the pod. We recommend you to check individual errors in the pod."
	// }

	r.logger.Info("checking for incident existence")
	if exists, err := r.redisClient.IncidentExists(ctx, incidentID); err != nil {
		return ctrl.Result{Requeue: true}, err
	} else if exists {
		r.logger.Info("incident already exists")
		// do not add duplicate events when possible
		r.logger.Info("fetching latest event for incident")
		latestEvent, err := r.redisClient.GetLatestEventForIncident(ctx, incidentID)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		} else if latestEvent != nil {
			// there is a special case when we get an "rpc error" followed by "Back-off pulling image"
			// we want to avoid logging this duo if it already has occurred
			if event.Reason == "Error while pulling image from container registry" &&
				latestEvent.Reason == event.Reason {
				// FIXME: a better way to check for this succession of events, perhaps?
				return ctrl.Result{}, nil
			}

			if strings.HasPrefix(event.Message, "back-off") {
				// FIXME: right now we just ignore these back-off errors
				return ctrl.Result{}, nil
			}

			if latestEvent.Reason == event.Reason && latestEvent.Message == event.Message {
				// since both the reason and the message are the same as the latest event for this incident,
				// we now check if this new event is an exact copy of the latest event
				newEvent := false

				if len(event.ContainerEvents) == len(latestEvent.ContainerEvents) {
					for containerName, containerEvent := range latestEvent.ContainerEvents {
						if _, ok := event.ContainerEvents[containerName]; !ok {
							// new container name so must be a new event
							newEvent = true
							break
						}

						newContainerEvent := event.ContainerEvents[containerName]

						if newContainerEvent.ExitCode != containerEvent.ExitCode ||
							newContainerEvent.Reason != containerEvent.Reason ||
							newContainerEvent.Message != containerEvent.Message {
							newEvent = true
							break
						}
					}
				} else {
					newEvent = true
				}

				if !newEvent {
					r.logger.Info("duplicate event")
					return ctrl.Result{}, nil
				}
			}
		}
	}

	r.logger.Info("fetching logs for containers")
	for containerName, containerEvent := range event.ContainerEvents {
		if containerEvent.LogID == "NOOP" {
			containerEvent.LogID = ""
			continue
		}

		logOptions := &corev1.PodLogOptions{
			TailLines: &maxTailLines,
			Previous:  true,
			Container: containerName,
		}

		req := r.KubeClient.
			CoreV1().
			Pods(instance.Namespace).
			GetLogs(instance.Name, logOptions)

		podLogs, err := req.Stream(ctx)
		if err != nil {
			r.logger.Error(err, "error streaming logs")
			return ctrl.Result{Requeue: true}, err
		}
		defer podLogs.Close()

		logs := new(bytes.Buffer)
		_, err = io.Copy(logs, podLogs)
		if err != nil {
			r.logger.Error(err, "unable to read logs")
			return ctrl.Result{Requeue: true}, err
		}

		strLogs := logs.String()

		if strLogs != "" { // logs can be empty
			if strings.Contains(strLogs, "unable to retrieve container logs") {
				// let us not add this unhelpful log message and completely ignore this event
				return ctrl.Result{}, nil
			}

			r.logger.Info("checking for duplicate logs", "incidentID", incidentID)

			duplicateLogs, err := r.redisClient.DuplicateLogs(ctx, incidentID, strLogs)
			if err != nil {
				r.logger.Error(err, "unable to check for duplicate logs")
				return ctrl.Result{Requeue: true}, err
			}

			if duplicateLogs {
				r.logger.Info("found duplicate logs", "incidentID", incidentID)
				return ctrl.Result{}, nil
			}

			logID, err := r.redisClient.AddLogs(ctx, incidentID, strLogs)
			if err != nil {
				r.logger.Error(err, "error adding new logs")
				return ctrl.Result{Requeue: true}, err
			}

			containerEvent.LogID = logID
		}
	}

	event.Message = getFilteredMessage(event.Message)

	for _, containerEvent := range event.ContainerEvents {
		containerEvent.Message = getFilteredMessage(containerEvent.Message)
	}

	r.logger.Info("adding event to incident")
	err = r.redisClient.AddEventToIncident(ctx, incidentID, event)
	if err != nil && strings.Contains(err.Error(), "max event count") {
		r.logger.Error(err, "max events reached for incident")
		return ctrl.Result{}, nil
	} else if err != nil {
		r.logger.Error(err, "error adding event to incident")
		return ctrl.Result{Requeue: true}, err
	}

	if isNewIncident {
		r.notifyNewIncident(ctx, incidentID)
	}

	return ctrl.Result{}, nil
}

func (r *PodReconciler) notifyNewIncident(ctx context.Context, incidentID string) {

}

func getFilteredMessage(message string) string {
	regex := regexp.MustCompile("failed to start container \".*?\"")
	matches := regex.FindStringSubmatch(message)

	filteredMsg := ""

	if len(matches) > 0 {
		containerName := strings.Split(matches[0], "\"")[1]
		filteredMsg = "In container \"" + containerName + "\": "
	}

	regex = regexp.MustCompile("starting container process caused:.*$")
	matches = regex.FindStringSubmatch(message)

	if len(matches) > 0 {
		filteredMsg += strings.TrimPrefix(matches[0], "starting container process caused: ")
	}

	if filteredMsg == "" {
		return message
	}

	return filteredMsg
}

func getFilteredReason(reason string) string {
	// refer: https://stackoverflow.com/a/57886025
	if reason == "CrashLoopBackOff" {
		return "Container is in a crash loop"
	} else if reason == "ImagePullBackOff" || reason == "ErrImagePull" {
		return "Error while pulling image from container registry"
	} else if reason == "OOMKilled" {
		return "Out-of-memory, resources exhausted"
	} else if reason == "Error" {
		return "Internal error"
	} else if reason == "ContainerCannotRun" {
		return "Container is unable to run due to internal error"
	} else if reason == "DeadlineExceeded" {
		return "Operation not completed in given timeframe"
	}

	return "Kubernetes error: " + reason
}

func (r *PodReconciler) getReasonAndMessage(instance *corev1.Pod, ownerType string) (string, string) {
	// since list is already sorted in place now, hence the first condition
	// is the latest, get its reason and message

	if len(instance.Spec.Containers) > 1 {
		latest := instance.Status.Conditions[0]
		r.logger.Info("multicontainer scenario", "latest", latest)

		// if pod is owned by a job, check in all containers
		if ownerType == "Job" {
			r.logger.Info("pod owned by a job. extracting from all container statuses of the pod")
			return r.extractMultipleContainerStatuses(instance)
		}

		// check if a failing container is mentioned in the message
		container, ok := utils.ExtractErroredContainer(latest.Message)
		r.logger.Info("extracted errored containers", "container", container, "ok", ok)
		if ok {
			r.logger.Info("failing container in message")
			// extract message and reason from given container
			return r.extractFromContainerStatuses(container, instance)
		}

		return r.extractMultipleContainerStatuses(instance)
	}

	r.logger.Info("extracting details from container statuses")
	// since its a single container call with empty container name
	return r.extractFromContainerStatuses("", instance)
}

func (r *PodReconciler) extractMultipleContainerStatuses(instance *corev1.Pod) (string, string) {
	reasons := []string{}
	messages := []string{}

	for _, container := range instance.Status.ContainerStatuses {
		reason, message := r.extractFromContainerStatuses(container.Name, instance)

		reasons = append(reasons, fmt.Sprintf("Container: %s, Reason: %s", container.Name, reason))
		messages = append(messages, fmt.Sprintf("Container: %s, Reason: %s", container.Name, message))
	}

	return strings.Join(reasons, "\n"), strings.Join(messages, "\n")
}

func (r *PodReconciler) extractFromContainerStatuses(containerName string, instance *corev1.Pod) (string, string) {
	var state corev1.ContainerState

	if containerName == "" {
		// this is from a pod with single container
		state = instance.Status.ContainerStatuses[0].State
	} else {
		for _, status := range instance.Status.ContainerStatuses {
			if status.Name == containerName {
				state = status.State
			}
		}
	}

	if state.Running != nil {
		return "", fmt.Sprintf("Container started at: %s", state.Running.StartedAt.Format(time.RFC3339))
	}

	if state.Terminated != nil {
		return fmt.Sprintf("State: %s, reason: %s", "Terminated", state.Terminated.Reason), state.Terminated.Message
	}

	return fmt.Sprintf("State: %s, reason: %s", "Waiting", state.Waiting.Reason), state.Waiting.Message
}

func (r *PodReconciler) fetchReplicaSetOwner(ctx context.Context, req ctrl.Request) (*metav1.OwnerReference, error) {
	rs := &appsv1.ReplicaSet{}

	err := r.Client.Get(ctx, req.NamespacedName, rs)
	if err != nil {
		r.logger.Error(err, "cannot fetch replicaset object")
		return nil, err
	}

	// get replicaset owner
	owners := rs.ObjectMeta.OwnerReferences
	if len(owners) == 0 {
		r.logger.Info("no owner for teh replicaset", "replicaset name", req.NamespacedName)
		return nil, fmt.Errorf("no owner defined for replicaset")
	}

	owner := owners[0]
	return &owner, nil
}

func (r *PodReconciler) getOwnerDetails(ctx context.Context, req ctrl.Request, pod *corev1.Pod) (string, string) {
	owners := pod.ObjectMeta.OwnerReferences

	if len(owners) == 0 {
		r.logger.Info("no owners defined for the pod")
		return "", ""
	}

	// in case of multiple owners, take the first
	var owner *metav1.OwnerReference
	var err error

	owner = &owners[0]

	if owner.Kind == "ReplicaSet" {
		r.logger.Info("fetching owner for replicaset")
		owner, err = r.fetchReplicaSetOwner(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      owner.Name,
				Namespace: req.Namespace,
			},
		})

		if err != nil {
			r.logger.Error(err, "cannot fetch owner for replicaset")

			return "", ""
		}
	}

	return pod.Labels["app.kubernetes.io/instance"], owner.Kind
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}
