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

	"github.com/go-logr/logr"
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

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Processor processor.Interface

	logger logr.Logger
}

//+kubebuilder:rbac:groups=porter.run,resources=nodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=porter.run,resources=nodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=porter.run,resources=nodes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Node object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = log.FromContext(ctx)

	instance := &corev1.Node{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			r.logger.Error(err, "node deleted")

			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	r.logger.Info("received node event", "event details", instance)

	if instance.Status.Phase == corev1.NodePending ||
		instance.Status.Phase == corev1.NodeTerminated {
		// critical significance, trigger a notification
		panic("not implemented")
	}

	// check latest condition by sorting
	utils.NodeConditionsSorter(instance.Status.Conditions, true)

	if len(instance.Status.Conditions) == 0 {
		r.logger.Info("empty status conditions....reconciling")
		return ctrl.Result{Requeue: true}, nil
	}

	latestCondition := instance.Status.Conditions[0]

	if latestCondition.Status == corev1.ConditionFalse {
		// critical
		// r.addToQueue(ctx, req, instance, true)
	} else {
		// not critical
		r.logger.Info("not a critical event")
		r.Processor.EnqueueDetails(ctx, req.NamespacedName, &processor.EnqueueDetailOptions{
			NodeInstance: instance,
		})

		// r.addToQueue(ctx, req, instance, false)
	}

	return ctrl.Result{}, nil
}

func (r *NodeReconciler) addToQueue(ctx context.Context, req ctrl.Request, instance *corev1.Node, isCritical bool) {
	// reason, message := r.getReasonAndMessage(instance)
	eventDetails := &models.EventDetails{
		ResourceType: models.NodeResource,
		Name:         req.Name,
		Namespace:    req.Namespace,
		Message:      "",
		Reason:       "",
		Critical:     isCritical,
		Timestamp:    getTime(),
	}

	r.Processor.AddToWorkQueue(ctx, req.NamespacedName, eventDetails)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).

		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&corev1.Node{}).
		Complete(r)
}
