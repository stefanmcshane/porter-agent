package logstore

import (
	"fmt"
	"sort"
	"strings"
)

func LabelsMapToString(labels map[string]string, matcher string) string {
	lstrs := make([]string, 0, len(labels))

	for l, v := range labels {
		lstrs = append(lstrs, fmt.Sprintf("%s%s%q", l, matcher, v))
	}

	sort.Strings(lstrs)
	return fmt.Sprintf("{%s}", strings.Join(lstrs, ", "))
}
