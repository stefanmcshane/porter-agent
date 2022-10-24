package log

import (
	"net/http"
	"time"

	"github.com/porter-dev/porter-agent/api/server/config"
	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/pkg/logstore"
	"github.com/porter-dev/porter/api/server/shared"
	"github.com/porter-dev/porter/api/server/shared/apierrors"
)

type GetLogHandler struct {
	decoderValidator shared.RequestDecoderValidator
	resultWriter     shared.ResultWriter

	Config *config.Config
}

func NewGetLogHandler(config *config.Config) *GetLogHandler {
	return &GetLogHandler{
		resultWriter:     shared.NewDefaultResultWriter(config.Logger, config.Alerter),
		decoderValidator: shared.NewDefaultRequestDecoderValidator(config.Logger, config.Alerter),
		Config:           config,
	}
}

func (h *GetLogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &types.GetLogRequest{}

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

	lb := &logBuffer{}
	stopCh := make(chan struct{})

	err := h.Config.LogStore.Query(logstore.QueryOptions{
		Start:       *req.StartRange,
		End:         *req.EndRange,
		Limit:       uint32(req.Limit),
		SearchParam: req.SearchParam,
		Direction:   req.Direction,
		Labels: map[string]string{
			"pod":       req.PodSelector,
			"namespace": req.Namespace,
		},
		CustomSelectorSuffix: "event_store!=\"true\",container!=\"sidecar\"",
	}, lb, stopCh)

	if err != nil {
		apierrors.HandleAPIError(h.Config.Logger, h.Config.Alerter, w, r, apierrors.NewErrInternal(err), true)
		return
	}

	res := &types.GetLogResponse{
		Logs:                 lb.Lines,
		BackwardContinueTime: lb.BackwardContinueTime,
		ForwardContinueTime:  lb.ForwardContinueTime,
	}

	h.resultWriter.WriteResult(w, r, res)
}

type logBuffer struct {
	Lines                []types.LogLine
	BackwardContinueTime *time.Time
	ForwardContinueTime  *time.Time
}

func (l *logBuffer) Write(timestamp *time.Time, log string) error {
	if l.Lines == nil {
		l.Lines = make([]types.LogLine, 0)
	}

	if l.BackwardContinueTime == nil {
		l.BackwardContinueTime = timestamp
	}

	if l.ForwardContinueTime == nil {
		l.ForwardContinueTime = timestamp
	}

	l.Lines = append(l.Lines, types.LogLine{
		Timestamp: timestamp,
		Line:      log,
	})

	if timestamp.Before(*l.BackwardContinueTime) {
		l.BackwardContinueTime = timestamp
	}

	if timestamp.After(*l.ForwardContinueTime) {
		l.ForwardContinueTime = timestamp
	}

	return nil
}
