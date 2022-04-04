package consumer

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"context"

	"github.com/go-logr/logr"
	porterErrors "github.com/porter-dev/porter-agent/pkg/errors"
	"github.com/porter-dev/porter-agent/pkg/httpclient"
	"github.com/porter-dev/porter-agent/pkg/pulsar"
	"github.com/porter-dev/porter-agent/pkg/redis"
	"github.com/spf13/viper"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	redisHost    string
	redisPort    string
	maxTailLines int64
	porterHost   string
	porterPort   string
	porterToken  string
	clusterID    string
	projectID    string

	consumerLog = ctrl.Log.WithName("event-consumer")
)

func init() {
	viper.SetDefault("REDIS_HOST", "porter-redis-master")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("MAX_TAIL_LINES", int64(100))
	viper.SetDefault("PORTER_PORT", "80")
	viper.AutomaticEnv()

	redisHost = viper.GetString("REDIS_HOST")
	redisPort = viper.GetString("REDIS_PORT")
	maxTailLines = viper.GetInt64("MAX_TAIL_LINES")

	porterPort = viper.GetString("PORTER_PORT")
	porterHost = getStringOrDie("PORTER_HOST")
	porterToken = getStringOrDie("PORTER_TOKEN")
	clusterID = getStringOrDie("CLUSTER_ID")
	projectID = getStringOrDie("PROJECT_ID")

}

type EventConsumer struct {
	redisClient *redis.Client
	httpClient  *httpclient.Client
	pulsar      *pulsar.Pulsar
	context     context.Context
	consumerLog logr.Logger
}

func getStringOrDie(key string) string {
	value := viper.GetString(key)

	if value == "" {
		panic(fmt.Errorf("empty %s", key))
		// consumerLog.Error(fmt.Errorf("empty %s", key), fmt.Sprintf("%s must not be empty", key))
		// os.Exit(1)
	}

	return value
}

func NewEventConsumer(timePeriod int, timeUnit time.Duration, ctx context.Context) *EventConsumer {
	return &EventConsumer{
		redisClient: redis.NewClient(redisHost, redisPort, "", "", redis.PODSTORE, maxTailLines),
		httpClient:  httpclient.NewClient(fmt.Sprintf("%s:%s", porterHost, porterPort), porterToken),
		pulsar:      pulsar.NewPulsar(timePeriod, timeUnit),
		context:     ctx,
		consumerLog: consumerLog,
	}
}

func (e *EventConsumer) Start() {
	e.consumerLog.Info("Starting event consumer")
	for range e.pulsar.Pulsate() {
		value, score, err := e.redisClient.GetItemFromPendingQueue(e.context)
		if err != nil {
			// log the error and continue
			if !errors.Is(err, porterErrors.NoPendingItemError) {
				e.consumerLog.Error(err, "cannot get pending item from store")
			}
			continue
		}

		payload := string(value)
		incidentID := ""
		newIncident := false

		if strings.HasPrefix(payload, "new:") {
			newIncident = true
			incidentID = strings.TrimPrefix(payload, "new:")
		} else if strings.HasPrefix(payload, "resolved:") {
			incidentID = strings.TrimPrefix(payload, "resolved:")
		}

		e.consumerLog.Info("doing HTTP post", "payload", payload)

		if newIncident {
			if err = e.doHTTPPostNotifyNew(incidentID); err != nil {
				// log error
				e.consumerLog.Error(err, "error sending HTTP request to porter server for new incident", "payload", payload)

				// requeue the object into the work queue
				if !strings.Contains(err.Error(), "non-existent incident") {
					err := e.redisClient.RequeueItemWithScore(e.context, value, score)
					if err != nil {
						// log error and continue
						e.consumerLog.Error(err, "error requeuing item in store with score", "payload", payload)
						continue
					}
				}
			}
		} else {
			if err = e.doHTTPPostNotifyResolved(incidentID); err != nil {
				// log error
				e.consumerLog.Error(err, "error sending HTTP request to porter server for resolved incident", "payload", payload)

				if !strings.Contains(err.Error(), "non-existent incident") {
					// requeue the object into the work queue
					err := e.redisClient.RequeueItemWithScore(e.context, value, score)
					if err != nil {
						// log error and continue
						e.consumerLog.Error(err, "error requeuing item in store with score", "payload", payload)
						continue
					}
				}
			}
		}
	}
}

func (e *EventConsumer) doHTTPPostNotifyNew(incidentID string) error {
	e.consumerLog.Info("notify new", "incidentID", incidentID)

	incident, err := e.redisClient.GetIncidentDetails(e.context, incidentID)
	if err != nil {
		e.consumerLog.Error(err, "error sending http request for new incident")
		return err
	}

	_, err = e.httpClient.Post(fmt.Sprintf("/api/projects/%s/clusters/%s/incidents/notify_new", projectID, clusterID), incident)

	if err != nil {
		// log and return error
		e.consumerLog.Error(err, "error sending http request for new incident", "incidentID", incidentID)
		return err
	}

	return nil
}

func (e *EventConsumer) doHTTPPostNotifyResolved(incidentID string) error {
	e.consumerLog.Info("notify resolved", "incidentID", incidentID)

	incident, err := e.redisClient.GetIncidentDetails(e.context, incidentID)
	if err != nil {
		e.consumerLog.Error(err, "error sending http request for new incident")
		return err
	}

	_, err = e.httpClient.Post(fmt.Sprintf("/api/projects/%s/clusters/%s/incidents/notify_resolved", projectID, clusterID), incident)

	if err != nil {
		// log and return error
		e.consumerLog.Error(err, "error sending http request for resolved incident", "incidentID", incidentID)
		return err
	}

	return nil
}
