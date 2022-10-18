package k8s_event

import (
	"net/http"
	"time"

	"github.com/porter-dev/porter-agent/api/server/config"
	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/pkg/logstore"
	"github.com/porter-dev/porter/api/server/shared"
	"github.com/porter-dev/porter/api/server/shared/apierrors"
)

type GetEventHandler struct {
	decoderValidator shared.RequestDecoderValidator
	resultWriter     shared.ResultWriter

	Config *config.Config
}

func NewGetEventHandler(config *config.Config) *GetEventHandler {
	return &GetEventHandler{
		resultWriter:     shared.NewDefaultResultWriter(config.Logger, config.Alerter),
		decoderValidator: shared.NewDefaultRequestDecoderValidator(config.Logger, config.Alerter),
		Config:           config,
	}
}

func (h *GetEventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &types.GetKubernetesEventRequest{}

	if ok := h.decoderValidator.DecodeAndValidate(w, r, req); !ok {
		return
	}

	if req.StartRange == nil {
		days29 := time.Now().Add(-29 * 24 * time.Hour)
		req.StartRange = &days29
	}

	if req.EndRange == nil {
		now := time.Now()
		req.EndRange = &now
	}

	eb := &eventBuffer{}
	stopCh := make(chan struct{})

	err := h.Config.LogStore.Query(logstore.QueryOptions{
		Start: *req.StartRange,
		End:   *req.EndRange,
		Limit: uint32(req.Limit),
		Labels: map[string]string{
			"pod":       req.PodSelector,
			"namespace": req.Namespace,
		},
		CustomSelectorSuffix: "event_store=\"true\"",
	}, eb, stopCh)

	if err != nil {
		apierrors.HandleAPIError(h.Config.Logger, h.Config.Alerter, w, r, apierrors.NewErrInternal(err), true)
		return
	}

	res := &types.GetKubernetesEventResponse{
		Events:       eb.Events,
		ContinueTime: eb.EarliestTimestamp,
	}

	h.resultWriter.WriteResult(w, r, res)
}

type eventBuffer struct {
	Events            []types.KubernetesEventLine
	EarliestTimestamp *time.Time
}

func (l *eventBuffer) Write(timestamp *time.Time, event string) error {
	if l.Events == nil {
		l.Events = make([]types.KubernetesEventLine, 0)
	}

	l.Events = append(l.Events, types.KubernetesEventLine{
		Timestamp: timestamp,
		Event:     event,
	})

	l.EarliestTimestamp = timestamp

	return nil
}
