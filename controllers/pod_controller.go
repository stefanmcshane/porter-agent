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
	"context"
	"fmt"
	"time"

	"github.com/porter-dev/porter-agent/pkg/models"
	"github.com/porter-dev/porter-agent/pkg/processor"
	"github.com/porter-dev/porter-agent/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Processor processor.Interface
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
	reqLogger := log.FromContext(ctx)

	instance := &corev1.Pod{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// must have been a delete event
			reqLogger.Info("pod deleted")

			// TODO: triggerNotify
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	// else we still have the object
	// we can log the current condition
	if instance.Status.Phase == corev1.PodFailed ||
		instance.Status.Phase == corev1.PodUnknown {
		// critical condition, must trigger a notification
		r.addToQueue(ctx, req, instance, true)
		return ctrl.Result{}, nil
	}

	// check ready condition
	// for _, condition := range instance.Status.Conditions {
	// 	if condition.Type == corev1.PodReady {
	// 		if condition.Status == corev1.ConditionFalse {
	// 			// pod is experiencing issues and is not
	// 			// in ready condition, hence trigger notification
	// 			r.triggerNotify(ctx, req, instance, true)
	// 			return ctrl.Result{}, nil
	// 		}
	// 	}
	// }

	// check latest condition by sorting
	utils.PodConditionsSorter(instance.Status.Conditions, true)

	// in case status conditions are not yet set for the pod
	// reconcile the event
	if len(instance.Status.Conditions) == 0 {
		reqLogger.Info("empty status conditions....reconciling")
		return ctrl.Result{Requeue: true}, nil
	}

	latestCondition := instance.Status.Conditions[0]

	if latestCondition.Status == corev1.ConditionFalse {
		// latest condition status is false, hence trigger notification
		r.addToQueue(ctx, req, instance, true)
		return ctrl.Result{}, nil
	}

	if (instance.Status.Phase != corev1.PodPending) &&
		(instance.Status.Phase != corev1.PodFailed) {
		// trigger with critical false
		r.addToQueue(ctx, req, instance, false)

		// normal event, fetch and enqueue latest logs
		reqLogger.Info("processing logs for pod", "status", instance.Status)

		if len(instance.Spec.Containers) > 1 {
			r.Processor.EnqueueDetails(ctx, req.NamespacedName, &processor.EnqueueDetailOptions{
				ContainerNamesToFetchLogs: []string{instance.Spec.Containers[0].Name},
			})
		} else {
			r.Processor.EnqueueDetails(ctx, req.NamespacedName, &processor.EnqueueDetailOptions{})
		}
	}

	return ctrl.Result{}, nil
}

func (r *PodReconciler) addToQueue(ctx context.Context, req ctrl.Request, instance *corev1.Pod, isCritical bool) {
	reason, message := r.getReasonAndMessage(instance)
	r.Processor.AddToWorkQueue(ctx, req.NamespacedName,
		models.EventDetails{
			ResourceType: models.PodResource,
			Name:         req.Name,
			Namespace:    req.Namespace,
			Message:      message,
			Reason:       reason,
			Critical:     isCritical,
			Timestamp:    getTime(),
			Phase:        string(instance.Status.Phase),
			Status:       fmt.Sprintf("Type: %s, Status: %s", instance.Status.Conditions[0].Type, instance.Status.Conditions[0].Status),
		})
}

func (r *PodReconciler) getReasonAndMessage(instance *corev1.Pod) (string, string) {
	// for _, condition := range instance.Status.Conditions {
	// 	if condition.Type == corev1.PodReady &&
	// 		condition.Status != corev1.ConditionTrue {
	// 		return condition.Reason, condition.Message
	// 	}
	// }

	// since list is already sorted in place now, hence the first condition
	// is the latest, get its reason and message

	latest := instance.Status.Conditions[0]
	return latest.Reason, latest.Message
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}

func getTime() string {
	return time.Now().Format(time.RFC3339)
}
