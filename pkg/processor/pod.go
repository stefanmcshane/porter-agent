package processor

type Interface interface {
	// Used in case of normal events to store and update logs
	EnqueueWithLogLines(object interface{}, loglines []string)
	// to trigger actual request for porter server in case of
	// a Delete or Failed/Unknown Phase
	TriggerNotifyForFatalEvent(object interface{})
}
