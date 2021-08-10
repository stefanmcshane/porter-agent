package processor

import (
	"context"

	"github.com/porter-dev/porter-agent/pkg/models"
	"k8s.io/apimachinery/pkg/types"
)

type Interface interface {
	// Used in case of normal events to store and update logs
	EnqueueWithLogLines(context context.Context, object types.NamespacedName)
	// to trigger actual request for porter server in case of
	// a Delete or Failed/Unknown Phase
	TriggerNotifyForEvent(context context.Context, object types.NamespacedName, details models.EventDetails)
}
