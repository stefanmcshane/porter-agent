package main

import (
	"flag"
	"log"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter/api/server/shared/config/env"

	"github.com/gin-gonic/gin"
	"github.com/porter-dev/porter-agent/controllers"
	"github.com/porter-dev/porter-agent/internal/adapter"
	"github.com/porter-dev/porter-agent/pkg/consumer"
	"github.com/porter-dev/porter-agent/pkg/incident"
	//+kubebuilder:scaffold:imports
)

var (
	scheme        = runtime.NewScheme()
	setupLog      = ctrl.Log.WithName("setup")
	eventConsumer *consumer.EventConsumer

	httpServer *gin.Engine
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8000", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "5731d595.porter.run",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	kubeClient := kubernetes.NewForConfigOrDie(mgr.GetConfig())

	detector := &incident.IncidentDetector{
		KubeClient: kubeClient,
		// TODO: don't hardcode to 1.20
		KubeVersion: incident.KubernetesVersion_1_20,
	}

	// eventController := controllers.EventController{
	// 	KubeClient: kubeClient,
	// 	// TODO: don't hardcode to 1.20
	// 	KubeVersion:      incident.KubernetesVersion_1_20,
	// 	IncidentDetector: detector,
	// }

	// eventController.Start()

	podController := controllers.PodController{
		KubeClient: kubeClient,
		// TODO: don't hardcode to 1.20
		KubeVersion:      incident.KubernetesVersion_1_20,
		IncidentDetector: detector,
	}

	podController.Start()

	db, err := adapter.New(&env.DBConf{})

	if err != nil {
		log.Fatalf("error opening connection to DB: %v\n", err)
	}

	go cleanupEventCache(db)

	// for {
	// 	pods, err := kubeClient.CoreV1().Pods("porter-agent-system").List(
	// 		context.Background(), v1.ListOptions{
	// 			LabelSelector: "app.kubernetes.io/name=redis",
	// 		},
	// 	)

	// 	if err == nil && len(pods.Items) > 0 {
	// 		running := false

	// 		for _, pod := range pods.Items {
	// 			if pod.Status.Phase == corev1.PodRunning {
	// 				running = true
	// 				break
	// 			}
	// 		}

	// 		if running {
	// 			break
	// 		}
	// 	}

	// 	setupLog.Info("waiting for redis ...")
	// 	time.Sleep(time.Second * 2)
	// }

	// if err = (&controllers.PodReconciler{
	// 	Client:     mgr.GetClient(),
	// 	Scheme:     mgr.GetScheme(),
	// 	KubeClient: kubeClient,
	// 	PodFilter:  utils.NewAgentPodFilter(kubeClient),
	// }).SetupWithManager(mgr); err != nil {
	// 	setupLog.Error(err, "unable to create controller", "controller", "Pod")
	// 	os.Exit(1)
	// }
	// //+kubebuilder:scaffold:builder

	// if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
	// 	setupLog.Error(err, "unable to set up health check")
	// 	os.Exit(1)
	// }
	// if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
	// 	setupLog.Error(err, "unable to set up ready check")
	// 	os.Exit(1)
	// }

	// // create the event consumer
	// setupLog.Info("creating event consumer")
	// eventConsumer = consumer.NewEventConsumer(50, time.Millisecond, context.TODO())

	// setupLog.Info("starting event consumer")
	// go eventConsumer.Start()

	// setupLog.Info("starting HTTP server")
	// httpServer = routes.NewRouter()
	// go httpServer.Run(":10001")

	// go func() {
	// 	// every 5 minutes, we check for deleted pods
	// 	// if a pod that was part of an active incident
	// 	// was deleted, we set the pod's status to resolved
	// 	for {
	// 		time.Sleep(time.Minute * 5)
	// 		controllers.ProcessDeletedPods()
	// 	}
	// }()

	// setupLog.Info("starting manager")
	// if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
	// 	setupLog.Error(err, "problem running manager")
	// 	os.Exit(1)
	// }
}

func cleanupEventCache(db *gorm.DB) {
	for {
		log.Println("cleaning old event cache entries from DB")

		var olderCache []*models.EventCache

		if err := db.Model(&models.EventCache{}).Where("timestamp <= ?", time.Now().Add(-time.Hour)).Find(&olderCache).Error; err == nil {
			for _, cache := range olderCache {
				if err := db.Delete(cache).Error; err != nil {
					log.Printf("error deleting old event cache with ID: %d. Error: %v\n", cache.ID, err)
				}
			}

			log.Println("old event cache entries deleted from DB")
		} else {
			log.Printf("error querying for older event cache DB entries: %v\n", err)
		}

		time.Sleep(time.Hour)
	}
}
