package ms_include

import (
	"regexp"
)

var urlRegex = regexp.MustCompile(`^[^/:]+:`)

func isFilePath(tgt string) bool {
	return !urlRegex.MatchString(tgt)
}

