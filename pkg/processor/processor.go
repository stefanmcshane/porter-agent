package processor

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
)

type Interface interface {
	// Used in case of normal events to store and update logs
	EnqueueWithLogLines(context context.Context, object types.NamespacedName)
	// to trigger actual request for porter server in case of
	// a Delete or Failed/Unknown Phase
	TriggerNotifyForFatalEvent(object types.NamespacedName, details map[string]interface{})
}
