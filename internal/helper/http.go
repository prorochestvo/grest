package helper

import (
	"regexp"
)

func HttpPathTrim(path string) string {
	path = regexp.MustCompile(`^/+`).ReplaceAllString(path, "")
	path = regexp.MustCompile(`/+$`).ReplaceAllString(path, "")
	return path
}

func HttpPathID(path string) (id, name, pattern string) {
	rx := regexp.MustCompile(`/{([^:]+):(.*)}$`)
	if tmp := rx.FindAllStringSubmatch(path, -1); tmp != nil && len(tmp) > 0 && len(tmp[0]) == 3 {
		id = tmp[0][0]
		name = tmp[0][1]
		pattern = tmp[0][2]
	}
	return
}
