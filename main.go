package main

import (
	"flag"
	"fmt"
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
	"github.com/joeshaw/envdecode"
	"github.com/porter-dev/porter-agent/controllers"
	"github.com/porter-dev/porter-agent/internal/adapter"
	"github.com/porter-dev/porter-agent/internal/repository"
	"github.com/porter-dev/porter-agent/pkg/incident"
	"github.com/porter-dev/porter-agent/pkg/pulsar"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	httpServer *gin.Engine
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

type EnvDecoderConf struct {
	DBConf env.DBConf
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

	var envDecoderConf EnvDecoderConf = EnvDecoderConf{}

	if err := envdecode.StrictDecode(&envDecoderConf); err != nil {
		setupLog.Error(err, "unable to decode env vars")
		os.Exit(1)
	}

	// create database connection through adapter
	db, err := adapter.New(&envDecoderConf.DBConf)

	if err != nil {
		setupLog.Error(err, "unable to create gorm db connection")
		os.Exit(1)
	}

	if err := repository.AutoMigrate(db, true); err != nil {
		setupLog.Error(err, "auto migration failed")
		os.Exit(1)
	}

	go cleanupEventCache(db)

	repo := repository.NewRepository(db)

	kubeClient := kubernetes.NewForConfigOrDie(mgr.GetConfig())

	detector := &incident.IncidentDetector{
		KubeClient: kubeClient,
		// TODO: don't hardcode to 1.20
		KubeVersion: incident.KubernetesVersion_1_20,
		Repository:  repo,
	}

	resolver := &incident.IncidentResolver{
		KubeClient: kubeClient,
		// TODO: don't hardcode to 1.20
		KubeVersion: incident.KubernetesVersion_1_20,
		Repository:  repo,
	}

	// trigger resolver through pulsar
	go func() {
		p := pulsar.NewPulsar(1, time.Minute) // pulse every 1 minute

		for range p.Pulsate() {
			err := resolver.Run()

			if err != nil {
				fmt.Println("pulsar error:", err)
			}
		}
	}()

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
