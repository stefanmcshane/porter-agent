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

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/gin-gonic/gin"
	"github.com/porter-dev/porter-agent/controllers"
	"github.com/porter-dev/porter-agent/pkg/consumer"
	"github.com/porter-dev/porter-agent/pkg/processor"
	"github.com/porter-dev/porter-agent/pkg/redis"
	"github.com/porter-dev/porter-agent/pkg/server/routes"
	"github.com/spf13/viper"
	//+kubebuilder:scaffold:imports
)

var (
	scheme        = runtime.NewScheme()
	setupLog      = ctrl.Log.WithName("setup")
	eventConsumer *consumer.EventConsumer

	httpServer *gin.Engine

	redisHost    string
	redisPort    string
	maxTailLines int64
)

func init() {
	viper.SetDefault("REDIS_HOST", "porter-redis-master")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("MAX_TAIL_LINES", int64(100))
	viper.AutomaticEnv()

	redisHost = viper.GetString("REDIS_HOST")
	redisPort = viper.GetString("REDIS_PORT")
	maxTailLines = viper.GetInt64("MAX_TAIL_LINES")

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

	redisClient := redis.NewClient(redisHost, redisPort, "", "", redis.PODSTORE, maxTailLines)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if firstRun, err := redisClient.IsFirstRun(ctx); err != nil {
		setupLog.Error(err, "error checking for first run of agent")
		os.Exit(1)
	} else if firstRun {
		err = redisClient.SetAgentCreationTimestamp(ctx)
		if err != nil {
			setupLog.Error(err, "error setting agent creation timestamp")
			os.Exit(1)
		}
	}

	if err = (&controllers.PodReconciler{
		Client:      mgr.GetClient(),
		Scheme:      mgr.GetScheme(),
		RedisClient: redisClient,
		KubeClient:  kubernetes.NewForConfigOrDie(mgr.GetConfig()),
		Processor:   processor.NewPodEventProcessor(mgr.GetConfig()),
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

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
