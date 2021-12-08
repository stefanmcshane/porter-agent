package utils

import (
	"regexp"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NamespaceFilter(filter map[string]bool) func(client.Object) bool {
	return func(object client.Object) bool {
		namespace := object.GetNamespace()
		if _, ok := filter[namespace]; ok {
			// ignore event from this namespace
			return false
		}

		return true
	}
}

func ExtractErroredContainer(msg string) (string, bool) {
	re := regexp.MustCompile(`\[([\w|-]+)\]`)

	if !re.MatchString(msg) {
		return "", false
	}

	match := re.FindStringSubmatch(msg)

	return match[1], true
}
