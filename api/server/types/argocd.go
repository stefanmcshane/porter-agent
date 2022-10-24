package types

// ArgoCDResourceHook represents the received event from an ArgoCD resource hook
type ArgoCDResourceHook struct {
	Application          string `json:"application"`
	ApplicationNamespace string `json:"namespace"`
	Status               string `json:"status"`
	Author               string `json:"author"`
	Timestamp            string `json:"timestamp"`
}
