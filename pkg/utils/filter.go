package utils

import "sigs.k8s.io/controller-runtime/pkg/client"

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
