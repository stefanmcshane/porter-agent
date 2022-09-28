package incident

import (
	"fmt"
	"regexp"

	"github.com/porter-dev/porter-agent/pkg/event"
)

// This file enumerates well-known event messages as regexes from pod controllers
type KubernetesVersion string

const (
	KubernetesVersion_1_20 KubernetesVersion = "1.20"
	KubernetesVersion_1_21 KubernetesVersion = "1.21"
	KubernetesVersion_1_22 KubernetesVersion = "1.22"
)

const RFC1123Name = `[a-z0-9]([-a-z0-9]*[a-z0-9])`

type ReconciliationLoopName string

const (
	LivenessProbeLoop ReconciliationLoopName = "liveness_probe"
)

type EventMatch struct {
	Summary         string
	DetailGenerator func(e *event.FilteredEvent) string

	SourceMatch  event.EventSource
	ReasonMatch  string
	MessageMatch *regexp.Regexp
	MatchedLoops []ReconciliationLoopName

	// IsPrimaryCause refers to whether an event match is the primary cause for a reconciliation
	// loop, or simply a proximate cause. For example, an application which is continuously failing
	// its liveness probe may be emitting critical "BackOff" events which are proximate causes.
	IsPrimaryCause bool
}

var EventEnum map[KubernetesVersion][]EventMatch

func init() {
	EventEnum = make(map[KubernetesVersion][]EventMatch)

	// Kubernetes 1.20 event matches
	eventMatch1_20 := make([]EventMatch, 0)

	// eventMatch1_20 = append(eventMatch1_20, EventMatch{
	// 	ReasonMatch: "CrashLoopBackOff",
	// })

	// eventMatch1_20 = append(eventMatch1_20, EventMatch{
	// 	ReasonMatch: "OOMKilled",
	// })

	// eventMatch1_20 = append(eventMatch1_20, EventMatch{
	// 	ReasonMatch: "StartError",
	// })

	// eventMatch1_20 = append(eventMatch1_20, EventMatch{
	// 	ReasonMatch: "InvalidImageName",
	// })

	// eventMatch1_20 = append(eventMatch1_20, EventMatch{
	// 	ReasonMatch: "Error",
	// })

	// eventMatch1_20 = append(eventMatch1_20, EventMatch{
	// 	ReasonMatch: "ContainerCannotRun",
	// })

	eventMatch1_20 = append(eventMatch1_20, EventMatch{
		SourceMatch: event.K8sEvent,
		Summary:     "The application is failing its health check",
		DetailGenerator: func(e *event.FilteredEvent) string {
			return "Your application was restarted because it failing its liveness health check. You can configure the liveness health check from the Advanced tab of your application settings."
		},
		ReasonMatch: "Killing",
		MessageMatch: regexp.MustCompile(
			fmt.Sprintf(`Container %s failed liveness probe, will be restarted`, RFC1123Name),
		),
		IsPrimaryCause: true,
	})

	eventMatch1_20 = append(eventMatch1_20, EventMatch{
		SourceMatch: event.K8sEvent,
		Summary:     "The application was restarted",
		DetailGenerator: func(e *event.FilteredEvent) string {
			return fmt.Sprintf("The application is stuck in a restart loop")
		},
		ReasonMatch:    "BackOff",
		MessageMatch:   regexp.MustCompile("Back-off restarting failed container"),
		IsPrimaryCause: false,
	})

	eventMatch1_20 = append(eventMatch1_20, EventMatch{
		SourceMatch: event.Pod,
		Summary:     "The application exited with a non-zero exit code",
		DetailGenerator: func(e *event.FilteredEvent) string {
			return fmt.Sprintf("The application restarted with exit code %d", e.ExitCode)
		},
		ReasonMatch:    "ApplicationError",
		MessageMatch:   regexp.MustCompile("Back-off restarting failed container"),
		IsPrimaryCause: true,
	})

	eventMatch1_20 = append(eventMatch1_20, EventMatch{
		SourceMatch: event.Pod,
		Summary:     "The application exited with a non-zero exit code",
		DetailGenerator: func(e *event.FilteredEvent) string {
			return fmt.Sprintf("The application restarted with exit code %d", e.ExitCode)
		},
		ReasonMatch:    "Error",
		MessageMatch:   regexp.MustCompile(".*"),
		IsPrimaryCause: true,
	})

	eventMatch1_20 = append(eventMatch1_20, EventMatch{
		SourceMatch: event.Pod,
		Summary:     "The application ran out of memory",
		DetailGenerator: func(e *event.FilteredEvent) string {
			return fmt.Sprintf("Your application ran out of memory. Reduce the amount of memory your application is consuming or bump up its memory limit from the Resources tab")
		},
		ReasonMatch:    "OOMKilled",
		MessageMatch:   regexp.MustCompile(".*"),
		IsPrimaryCause: true,
	})

	eventMatch1_20 = append(eventMatch1_20, EventMatch{
		SourceMatch: event.Pod,
		Summary:     "The application has an invalid image",
		DetailGenerator: func(e *event.FilteredEvent) string {
			return fmt.Sprintf("Your application cannot pull from the image registry.")
		},
		ReasonMatch:    "ImagePullBackOff",
		MessageMatch:   regexp.MustCompile(".*"),
		IsPrimaryCause: true,
	})

	EventEnum[KubernetesVersion_1_20] = eventMatch1_20
}

func GetEventMatchFromEvent(k8sVersion KubernetesVersion, filteredEvent *event.FilteredEvent) *EventMatch {
	for _, candidate := range EventEnum[k8sVersion] {
		if candidate.SourceMatch == filteredEvent.Source && candidate.ReasonMatch == filteredEvent.KubernetesReason {
			if candidate.MessageMatch.Match([]byte(filteredEvent.KubernetesMessage)) {
				return &candidate
			}
		}
	}

	return nil
}
