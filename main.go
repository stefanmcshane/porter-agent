package main

import (
	"context"
	"flag"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/gin-gonic/gin"
	"github.com/porter-dev/porter-agent/controllers"
	"github.com/porter-dev/porter-agent/pkg/consumer"
	"github.com/porter-dev/porter-agent/pkg/server/routes"
	"github.com/porter-dev/porter-agent/pkg/utils"
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

	// first check if the redis server is running and wait for it if needed
	kubeClient := kubernetes.NewForConfigOrDie(mgr.GetConfig())
	for {
		pods, err := kubeClient.CoreV1().Pods("porter-agent-system").List(
			context.Background(), v1.ListOptions{
				LabelSelector: "app.kubernetes.io/name=redis",
			},
		)

		if err == nil && len(pods.Items) > 0 {
			running := false

			for _, pod := range pods.Items {
				if pod.Status.Phase == corev1.PodRunning {
					running = true
					break
				}
			}

			if running {
				break
			}
		}

		setupLog.Info("waiting for redis ...")
		time.Sleep(time.Second * 2)
	}

	if err = (&controllers.PodReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		KubeClient: kubeClient,
		PodFilter:  utils.NewAgentPodFilter(kubeClient),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Pod")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	// create the event consumer
	setupLog.Info("creating event consumer")
	eventConsumer = consumer.NewEventConsumer(50, time.Millisecond, context.TODO())

	setupLog.Info("starting event consumer")
	go eventConsumer.Start()

	setupLog.Info("starting HTTP server")
	httpServer = routes.NewRouter()
	go httpServer.Run(":10001")

	go func() {
		// every 5 minutes, we check for deleted pods
		// if a pod that was part of an active incident
		// was deleted, we set the pod's status to resolved
		for {
			time.Sleep(time.Minute * 5)
			controllers.ProcessDeletedPods()
		}
	}()

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
