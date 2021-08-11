package consumer

import (
	"encoding/json"
	"errors"
	"io"
	"time"

	"context"

	"github.com/go-logr/logr"
	porterErrors "github.com/porter-dev/porter-agent/pkg/errors"
	"github.com/porter-dev/porter-agent/pkg/httpclient"
	"github.com/porter-dev/porter-agent/pkg/models"
	"github.com/porter-dev/porter-agent/pkg/pulsar"
	"github.com/porter-dev/porter-agent/pkg/redis"
	ctrl "sigs.k8s.io/controller-runtime"
)

type EventConsumer struct {
	redisClient *redis.Client
	httpClient  *httpclient.Client
	pulsar      *pulsar.Pulsar
	context     context.Context
	consumerLog logr.Logger
}

func NewEventConsumer(timePeriod int, timeUnit time.Duration, ctx context.Context) *EventConsumer {
	return &EventConsumer{
		redisClient: redis.NewClient("127.0.0.1", "6379", "", "", redis.PODSTORE, int64(100)),
		httpClient:  httpclient.NewClient("http://localhost:80", ""),
		pulsar:      pulsar.NewPulsar(timePeriod, timeUnit),
		context:     ctx,
		consumerLog: ctrl.Log.WithName("event consumer"),
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

		var payload *models.EventDetails
		err = json.Unmarshal(value, &payload)
		if err != nil {
			// log error and continue
			e.consumerLog.Error(err, "cannot unmarshal payload")
			continue
		}

		if payload.Critical {
			// include logs
			err := e.injectLogs(payload)
			if err != nil {
				e.consumerLog.Error(err, "unable to inject logs")
			}
		}

		if err = e.doHTTPPost(payload); err != nil {
			// log error
			e.consumerLog.Error(err, "error sending HTTP request to porter server")

			// requeue the object into the work queue
			err := e.redisClient.RequeueItemWithScore(e.context, value, score)
			if err != nil {
				// log error and continue
				e.consumerLog.Error(err, "error requeuing item in store with score")
				continue
			}
		}
	}
}

func (e *EventConsumer) injectLogs(payload *models.EventDetails) error {
	logs, err := e.redisClient.GetDetails(e.context, payload.ResourceType.String(), payload.Namespace, payload.Name)
	if err != nil {
		return err
	}

	payload.Data = logs
	return nil
}

// doHTTPRequest tries to do a POST http request to the porter
// server and returns an error if any. Its the responsibility of the caller
// to retain the object in case the requests fails or times out
func (e *EventConsumer) doHTTPPost(payload *models.EventDetails) error {
	response, err := e.httpClient.Post("/anything", payload)
	if err != nil {
		// log and return error
		e.consumerLog.Error(err, "error sending http request")
		return err
	}
	defer response.Body.Close()

	// log response and return
	e.consumerLog.Info("received response from server", "status", response.Status)
	body, err := io.ReadAll(response.Body)
	if err != nil {
		// log error and return
		e.consumerLog.Error(err, "error reading response body")
		return err
	}

	e.consumerLog.Info(string(body))
	return nil
}
