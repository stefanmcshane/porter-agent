package utils

import (
	"sort"

	corev1 "k8s.io/api/core/v1"
)

func PodConditionsSorter(conditions []corev1.PodCondition, reverse bool) {
	// logger := log.Log.WithName("pod-sorter")
	sort.SliceStable(conditions, func(i, j int) bool {
		t1 := conditions[i].LastTransitionTime.Time
		t2 := conditions[j].LastTransitionTime.Time

		if t1.Equal(t2) {
			// logger.Info("t1 and t2 are equal", string(conditions[i].Type), string(conditions[i].Status), string(conditions[j].Type), string(conditions[j].Status))
			// compare with condition status
			if conditions[i].Status != conditions[j].Status {
				if conditions[i].Status == corev1.ConditionTrue {
					// logger.Info("swapping")
					return reverse != true
				}
			}
		}

		return reverse != t1.Before(t2)
	})
}

func NodeConditionsSorter(conditions []corev1.NodeCondition, reverse bool) {
	sort.SliceStable(conditions, func(i, j int) bool {
		t1 := conditions[i].LastTransitionTime.Time
		t2 := conditions[j].LastTransitionTime.Time

		if t1.Equal(t2) {
			if conditions[i].Status != conditions[j].Status {
				if conditions[i].Status == corev1.ConditionTrue {
					// logger.Info("swapping")
					return reverse != true
				}
			}
		}

		return reverse != t1.Before(t2)
	})
}
