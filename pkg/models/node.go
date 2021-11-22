package models

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

type MarshallableNodeConditions []corev1.NodeCondition

func (m MarshallableNodeConditions) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}
