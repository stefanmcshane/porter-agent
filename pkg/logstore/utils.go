package logstore

import (
	"fmt"
	"sort"
	"strings"
)

func LabelsMapToString(labels map[string]string, matcher string, additionalQuery string) string {
	lstrs := make([]string, 0, len(labels))

	for l, v := range labels {
		lstrs = append(lstrs, fmt.Sprintf("%s%s%q", l, matcher, v))
	}

	sort.Strings(lstrs)

	if additionalQuery != "" {
		return fmt.Sprintf("{%s, %s}", strings.Join(lstrs, ", "), additionalQuery)
	}

	return fmt.Sprintf("{%s}", strings.Join(lstrs, ", "))
}

func ConstructSearch(labelsString, searchParam string) string {
	if searchParam == "" {
		return labelsString
	}

	return fmt.Sprintf("%s |~ \"(?i)%s\"", labelsString, searchParam)
}
