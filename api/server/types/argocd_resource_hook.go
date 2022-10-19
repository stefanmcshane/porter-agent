package types

// ArgoCDResourceHook represents the received event from an ArgoCD resource hook
type ArgoCDResourceHook struct {
	Test string `json:"test"`
}
