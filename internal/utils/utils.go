package utils

import (
	"fmt"
	"regexp"
)

func ReplacePostgresql(str string) string {
	regex, err := regexp.Compile(` postgresql://(.*) `)
	if err != nil {
		return str
	}

	secret := "******"
	matches := regex.FindStringSubmatch(str)
	if len(matches) >= 2 {
		replaced := regex.ReplaceAllString(str, fmt.Sprintf(" postgresql://%s ", secret))
		return replaced
	}

	return str
}

func ReplaceMinioSecret(str string) string {
	regex, err := regexp.Compile(`set ` + Alias + ` (.*) (.*) (.*) --api`)
	if err != nil {
		return str
	}

	secret := "******"
	matches := regex.FindStringSubmatch(str)
	if len(matches) >= 4 {
		// matches[1] = host
		// matches[2] = access key
		// matches[3] = secret key
		replaced := regex.ReplaceAllString(str, fmt.Sprintf("set %s %s --api", Alias, secret))
		return replaced
	}

	return str
}
