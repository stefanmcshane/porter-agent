package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
)

var porterHost string

type FilteredMessageResult struct {
	PodSummary        string
	PodDetails        string
	ContainerStatuses []*FilteredMessageContainerResult
}

type FilteredMessageContainerResult struct {
	ContainerName string
	Summary       string
	Details       string
}

type PodFilter interface {
	Filter(*corev1.Pod, bool, logr.Logger) *FilteredMessageResult
}

type AgentPodFilter struct{}

func init() {
	porterHost = viper.GetString("PORTER_HOST")
}

func NewAgentPodFilter() PodFilter {
	return &AgentPodFilter{}
}

func (f *AgentPodFilter) Filter(pod *corev1.Pod, isJob bool, log logr.Logger) *FilteredMessageResult {
	res := &FilteredMessageResult{}

	log.Info("filtering for", "pod", pod.Name)

	for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
		if isJob && (pod.Status.ContainerStatuses[i].Name == "sidecar" ||
			pod.Status.ContainerStatuses[i].Name == "cloud-sql-proxy") {
			continue
		}

		status := pod.Status.ContainerStatuses[i]
		containerResult := &FilteredMessageContainerResult{
			ContainerName: status.Name,
		}

		if status.State.Waiting != nil && status.State.Waiting.Reason != "" {
			if status.State.Waiting.Reason == "CrashLoopBackOff" {
				log.Info("pod is in CrashLoopBackOff")

				if status.LastTerminationState.Terminated != nil {
					if status.LastTerminationState.Terminated.Reason == "Error" {
						containerResult.Summary = fmt.Sprintf("The application exited with exit code %d",
							status.LastTerminationState.Terminated.ExitCode)
						containerResult.Details = fmt.Sprintf("The application exited with exit code %d. "+
							"Please see the list of exit codes in the Porter documentation: <docs-link>",
							status.LastTerminationState.Terminated.ExitCode)
					} else if status.LastTerminationState.Terminated.Reason == "OOMKilled" {
						containerResult.Summary = "The application was killed because it used too much memory"
						containerResult.Details = fmt.Sprintf("The application exceeded its memory limit of %s. ",
							pod.Spec.Containers[0].Resources.Limits.Memory().String())

						containerResult.Details += "Reduce the amount of memory your application is using or increase the memory limit - " +
							"see the docs here for more information: https://docs.porter.run/managing-applications/" +
							"application-troubleshooting#memory-usage"
					} else if status.LastTerminationState.Terminated.Reason == "ContainerCannotRun" {
						containerResult.Summary = "The application could not start running"
						containerResult.Details = getFilteredMessage(status.LastTerminationState.Terminated.Message)
					}
				}
			} else if status.State.Waiting.Reason == "ErrImagePull" ||
				status.State.Waiting.Reason == "ImagePullBackOff" {
				containerResult.Summary = "The image could not be pulled from the registry"
				containerResult.Details = fmt.Sprintf("The application was unable to pull image %s. "+
					"Please make sure you have linked this image registry to Porter by navigating to %s/"+
					"integrations/registry. See documentation for linking your registry here: "+
					"https://docs.porter.run/deploying-applications/deploying-from-docker-registry/linking-existing-registry",
					status.Image, porterHost)
			} else if status.State.Waiting.Reason == "InvalidImageName" {
				containerResult.Summary = "The image could not be pulled from the registry because the image URI is invalid"
				containerResult.Details = fmt.Sprintf("The specified image %s is not a valid image URI.", status.Image)
			}
		} else if status.State.Terminated != nil && status.State.Terminated.Reason != "" {
			if status.State.Terminated.Reason == "Error" {
				containerResult.Summary = fmt.Sprintf("The application exited with exit code %d",
					status.LastTerminationState.Terminated.ExitCode)
				containerResult.Details = fmt.Sprintf("The application exited with exit code %d. "+
					"Please see the list of exit codes in the Porter documentation: <docs-link>",
					status.LastTerminationState.Terminated.ExitCode)
			} else if status.State.Terminated.Reason == "OOMKilled" {
				containerResult.Summary = "The application was killed because it used too much memory"
				containerResult.Details = fmt.Sprintf("The application exceeded its memory limit of %s. ",
					pod.Spec.Containers[0].Resources.Limits.Memory().String())

				containerResult.Details += "Reduce the amount of memory your application is using or increase the memory limit - " +
					"see the docs here for more information: https://docs.porter.run/managing-applications/" +
					"application-troubleshooting#memory-usage"
			} else if status.State.Terminated.Reason == "ContainerCannotRun" {
				containerResult.Summary = "The application could not start running"
				containerResult.Details = getFilteredMessage(status.State.Terminated.Message)
			}
		} else if status.State.Terminated != nil && status.State.Terminated.Reason == "" {
			containerResult.Summary = fmt.Sprintf("The application exited with exit code %d",
				status.State.Terminated.ExitCode)
			containerResult.Details = fmt.Sprintf("The application exited with exit code %d. "+
				"Please see the list of exit codes in the Porter documentation: <docs-link>",
				status.State.Terminated.ExitCode)
		}

		if containerResult.Details != "" && containerResult.Summary != "" {
			res.ContainerStatuses = append(res.ContainerStatuses, containerResult)
		}
	}

	log.Info("filter result", "result", res)

	if len(res.ContainerStatuses) == 0 {
		return nil
	} else if len(res.ContainerStatuses) == 1 {
		res.PodSummary = res.ContainerStatuses[0].Summary
		res.PodDetails = res.ContainerStatuses[0].Details
	} else { // more than one container
		summary := ""
		details := ""

		for _, containerResult := range res.ContainerStatuses {
			summary += fmt.Sprintf("Container: %s. Summary: %s\n", containerResult.ContainerName, containerResult.Summary)
			details += fmt.Sprintf("Container: %s. Details: %s\n", containerResult.ContainerName, containerResult.Details)
		}

		res.PodSummary = summary
		res.PodDetails = details
	}

	return res
}

func getFilteredMessage(message string) string {
	regex := regexp.MustCompile("starting container process caused:.*$")
	matches := regex.FindStringSubmatch(message)

	if len(matches) > 0 {
		return strings.TrimPrefix(matches[0], "starting container process caused: ")
	}

	return message
}
