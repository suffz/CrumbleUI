package utils

import (
	"strings"
)

func HashEmailClean(email string) string {
	e := strings.Split(email, "@")[0] // stfu
	return e[0:len(e)/2] + "...."
}
