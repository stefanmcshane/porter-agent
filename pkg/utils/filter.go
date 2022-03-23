package utils

import (
	"regexp"
)

func ExtractErroredContainer(msg string) (string, bool) {
	re := regexp.MustCompile(`\[([\w|-]+)\]`)

	if !re.MatchString(msg) {
		return "", false
	}

	match := re.FindStringSubmatch(msg)

	return match[1], true
}
