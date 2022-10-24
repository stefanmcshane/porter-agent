package models

import "gorm.io/gorm"

// Alert stores additional data about alerts for an incident. This object was
// created in order to track the pods which have already been alerted for job runs,
// and which pods have not been alerted.
type Alert struct {
	gorm.Model

	// IncidentID is the name of the incident that this alert was triggered for.
	IncidentID uint

	// Summary is the incident summary that is associated with the incident. We store
	// this as part of the alerter configuration because we want to prevent job runs from
	// re-triggering the alerts if they have the exact same summary and pod name.
	Summary string

	// TriggeringPodName is the name of the pod which triggered this alert. This is
	// primarily applicable for job alerting.
	TriggeringPodName string
}
