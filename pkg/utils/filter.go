package utils

import (
	"regexp"
	"strings"
)

func ExtractErroredContainer(msg string) (string, bool) {
	re := regexp.MustCompile(`\[([\w|-]+)\]`)

	if !re.MatchString(msg) {
		return "", false
	}

	match := re.FindStringSubmatch(msg)

	return match[1], true
}

func GetFilteredMessage(message string) string {
	regex := regexp.MustCompile("failed to start container \".*?\"")
	matches := regex.FindStringSubmatch(message)

	filteredMsg := ""

	if len(matches) > 0 {
		containerName := strings.Split(matches[0], "\"")[1]
		filteredMsg = "In container \"" + containerName + "\": "
	}

	regex = regexp.MustCompile("starting container process caused:.*$")
	matches = regex.FindStringSubmatch(message)

	if len(matches) > 0 {
		filteredMsg += strings.TrimPrefix(matches[0], "starting container process caused: ")
	}

	if filteredMsg == "" {
		return message
	}

	return filteredMsg
}

func GetFilteredReason(reason string) string {
	// refer: https://stackoverflow.com/a/57886025
	if reason == "CrashLoopBackOff" {
		return "Container is in a crash loop"
	} else if reason == "ImagePullBackOff" || reason == "ErrImagePull" {
		return "Error while pulling image from container registry"
	} else if reason == "OOMKilled" {
		return "Out-of-memory, resources exhausted"
	} else if reason == "Error" {
		return "Internal error"
	} else if reason == "ContainerCannotRun" {
		return "Container is unable to run due to internal error"
	} else if reason == "DeadlineExceeded" {
		return "Operation not completed in given timeframe"
	}

	return "Kubernetes error: " + reason
}
