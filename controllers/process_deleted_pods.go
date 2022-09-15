package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/porter-dev/porter-agent/pkg/redis"
	"github.com/porter-dev/porter-agent/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	deletedPodsLogger = ctrl.Log.WithName("Deleted Pods")
)

func getClientset() (*kubernetes.Clientset, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()

	if err != nil {
		return nil, fmt.Errorf("Could not read in cluster config: %w", err)
	}

	// creates the clientset
	return kubernetes.NewForConfig(config)
}

func ProcessDeletedPods() {
	deletedPodsLogger.Info("Processing deleted pods")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	clientset, err := getClientset()

	if err != nil {
		deletedPodsLogger.Error(err, "error getting clientset for ProcessDeletedPods")
		return
	}

	redisClient := redis.NewClient(redisHost, redisPort, "", "", redis.PODSTORE, maxTailLines)

	incidentIDs, err := redisClient.GetAllActiveIncidents(ctx)

	if err != nil {
		deletedPodsLogger.Error(err, "error getting active incidents")
		return
	}

	deletedPodsLogger.Info(fmt.Sprintf("Found %d active incidents", len(incidentIDs)))

	for _, id := range incidentIDs {
		pods, err := redisClient.GetPodsForIncident(ctx, id)

		if err != nil {
			deletedPodsLogger.Error(err, "error getting active incidents")

			continue
		}

		deletedPodsLogger.Info(fmt.Sprintf("Found %d pods for incident %s", len(pods), id))

		incidentObj, _ := utils.NewIncidentFromString(id)

		for _, pod := range pods {
			deletedPodsLogger.Info(fmt.Sprintf("Checking if pod %s is deleted", pod))

			_, err := clientset.CoreV1().Pods(incidentObj.GetNamespace()).Get(ctx, pod, v1.GetOptions{})

			deletedPodsLogger.Info(fmt.Sprintf("Error getting pod %s: %v", pod, err))

			if err != nil && errors.IsNotFound(err) {
				// pod was deleted, so we should remove it from the incident
				redisClient.SetPodResolved(ctx, pod, id)
			}
		}
	}
}
