package controllers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/porter-dev/porter-agent/pkg/models"
	"github.com/porter-dev/porter-agent/pkg/redis"
	"github.com/porter-dev/porter-agent/pkg/utils"
	"github.com/spf13/viper"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
	Scheme *runtime.Scheme

	redisClient *redis.Client
	KubeClient  *kubernetes.Clientset
	PodFilter   utils.PodFilter

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

	porterReleaseName, ownerName, ownerKind, chartName := r.getOwnerDetails(ctx, req, instance)

	customFinalizer := "porter.run/agent-finalizer"

	finalizers := instance.Finalizers
	if ownerKind != "Job" {
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
				incidentID, err := r.redisClient.GetActiveIncident(ctx, porterReleaseName, instance.Namespace)
				if err == nil {
					r.redisClient.SetPodResolved(ctx, instance.Name, incidentID) // FIXME: make use of the error
				}

				// remove the finalizer
				var updatedFinalizers []string
				for _, fin := range finalizers {
					if fin != customFinalizer {
						updatedFinalizers = append(updatedFinalizers, fin)
					}
				}

				instance.SetFinalizers(updatedFinalizers)
				if err = r.Update(ctx, instance); err != nil {
					return ctrl.Result{Requeue: true}, fmt.Errorf("error removing custom finalizer: %w", err)
				}
			}

			return ctrl.Result{}, nil
		}
	}

	if ownerKind == "Deployment" {
		ignore, _ := r.canIgnoreMultipodDeployment(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      ownerName,
				Namespace: req.Namespace,
			},
		})

		if ignore {
			r.logger.Info("ignoring multipod deployment", "deployment", ownerName, "pod", instance.Name)
			// return ctrl.Result{}, nil
		}
	} else if ownerKind == "Job" {
		// we care only for the most recent pod for a job
		jobPods, err := r.KubeClient.CoreV1().Pods(instance.Namespace).List(
			ctx, metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app.kubernetes.io/instance=%s", porterReleaseName),
			},
		)

		if err != nil {
			r.logger.Error(err, "error fetching list of job pods", "job", porterReleaseName, "pod", instance.Name)
			return ctrl.Result{Requeue: true}, err
		}

		sort.SliceStable(jobPods.Items, func(i, j int) bool {
			return jobPods.Items[i].CreationTimestamp.After(jobPods.Items[j].CreationTimestamp.Time)
		})

		if jobPods.Items[0].Name != instance.Name {
			return ctrl.Result{}, nil
		}
	}

	r.logger.Info("creating container events")

	filteredMsgRes := r.PodFilter.Filter(instance, ownerKind == "Job")

	if filteredMsgRes == nil {
		incidentID, err := r.redisClient.GetActiveIncident(ctx, porterReleaseName, instance.Namespace)
		if err == nil {
			if ownerKind == "Job" {
				// since a job has one running pod at a time and here we know that it has run successfully
				r.redisClient.SetJobIncidentResolved(ctx, incidentID) // FIXME: make use of the error
			} else {
				r.redisClient.SetPodResolved(ctx, instance.Name, incidentID) // FIXME: make use of the error
			}
		}

		return ctrl.Result{}, nil // FIXME: better introspection to requeue here
	}

	containerEvents := make(map[string]*models.ContainerEvent)

	for _, filteredContainerRes := range filteredMsgRes.ContainerStatuses {
		containerEvents[filteredContainerRes.ContainerName] = &models.ContainerEvent{
			Name:    filteredContainerRes.ContainerName,
			Reason:  filteredContainerRes.Summary,
			Message: filteredContainerRes.Details,
		}
	}

	r.logger.Info("container events created.", "length", len(containerEvents))

	newIncident := false
	incidentID := ""

	exists, err := r.redisClient.ActiveIncidentExists(ctx, porterReleaseName, instance.Namespace)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	if exists {
		incidentID, err = r.redisClient.GetActiveIncident(ctx, porterReleaseName, instance.Namespace)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	} else {
		incidentID, err = r.redisClient.CreateActiveIncident(ctx, porterReleaseName, instance.Namespace)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}

		newIncident = true
	}

	r.logger.Info("active incident ID", "incidentID", incidentID)

	event := &models.PodEvent{
		ChartName:       chartName,
		PodName:         instance.Name,
		Namespace:       instance.Namespace,
		OwnerName:       porterReleaseName,
		OwnerType:       ownerKind,
		Timestamp:       time.Now().Unix(),
		Phase:           string(instance.Status.Phase),
		ContainerEvents: containerEvents,
		Reason:          filteredMsgRes.PodSummary,
		Message:         filteredMsgRes.PodDetails,
	}

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
		logOptions := &corev1.PodLogOptions{
			TailLines: &maxTailLines,
			Previous:  r.hasLastTerminatedState(instance, containerName),
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

	r.logger.Info("adding event to incident")
	err = r.redisClient.AddEventToIncident(ctx, incidentID, event, newIncident)
	if err != nil && strings.Contains(err.Error(), "max event count") {
		r.logger.Error(err, "max events reached for incident")
		return ctrl.Result{}, nil
	} else if err != nil {
		r.logger.Error(err, "error adding event to incident")
		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{}, nil
}

func (r *PodReconciler) hasLastTerminatedState(pod *corev1.Pod, containerName string) bool {
	for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
		if containerName == pod.Status.ContainerStatuses[i].Name {
			if pod.Status.ContainerStatuses[i].LastTerminationState.Waiting != nil ||
				pod.Status.ContainerStatuses[i].LastTerminationState.Terminated != nil {
				return true
			}
		}
	}

	return false
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

// returns the owner name, kind, chart name
func (r *PodReconciler) getOwnerDetails(ctx context.Context, req ctrl.Request, pod *corev1.Pod) (string, string, string, string) {
	owners := pod.ObjectMeta.OwnerReferences

	if len(owners) == 0 {
		r.logger.Info("no owners defined for the pod")
		return "", "", "", ""
	}

	// in case of multiple owners, take the first
	var owner *metav1.OwnerReference
	var err error

	owner = &owners[0]
	chartName := ""

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

			return "", "", "", ""
		}
	}

	chartName = r.getOwnerChartName(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      owner.Name,
			Namespace: req.Namespace,
		},
	}, owner.Kind == "Job")

	return pod.Labels["app.kubernetes.io/instance"], owner.Name, owner.Kind, chartName
}

func (r *PodReconciler) getOwnerChartName(ctx context.Context, req reconcile.Request, isJob bool) string {
	if isJob {
		job := &batchv1.Job{}

		err := r.Client.Get(ctx, req.NamespacedName, job)
		if err != nil {
			r.logger.Error(err, "cannot fetch job object")
			return ""
		}

		return job.Labels["helm.sh/chart"]
	}

	depl := &appsv1.Deployment{}

	err := r.Client.Get(ctx, req.NamespacedName, depl)
	if err != nil {
		r.logger.Error(err, "cannot fetch deployment object")
		return ""
	}

	return depl.Labels["helm.sh/chart"]
}

func getMaxUnavailable(deployment *appsv1.Deployment) int32 {
	if deployment.Spec.Strategy.Type != appsv1.RollingUpdateDeploymentStrategyType || *(deployment.Spec.Replicas) == 0 {
		return int32(0)
	}

	desired := *(deployment.Spec.Replicas)
	maxUnavailable := deployment.Spec.Strategy.RollingUpdate.MaxUnavailable

	unavailable, err := intstrutil.GetScaledValueFromIntOrPercent(intstrutil.ValueOrDefault(maxUnavailable, intstrutil.FromInt(0)), int(desired), false)

	if err != nil {
		return 0
	}

	return int32(unavailable)
}

func (r *PodReconciler) canIgnoreMultipodDeployment(ctx context.Context, req reconcile.Request) (bool, error) {
	depl := &appsv1.Deployment{}

	err := r.Client.Get(ctx, req.NamespacedName, depl)
	if err != nil {
		r.logger.Error(err, "cannot fetch deployment object")
		return false, err
	}

	minUnavailable := *(depl.Spec.Replicas) - getMaxUnavailable(depl)

	if minUnavailable > depl.Status.ReadyReplicas {
		return true, nil
	}

	return false, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}
