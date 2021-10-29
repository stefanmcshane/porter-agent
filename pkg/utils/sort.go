package utils

import (
	"sort"

	corev1 "k8s.io/api/core/v1"
)

func PodConditionsSorter(conditions []corev1.PodCondition, reverse bool) {
	sort.SliceStable(conditions, func(i, j int) bool {
		return reverse != conditions[i].LastTransitionTime.Time.
			Before(conditions[j].LastTransitionTime.Time)
	})
}
