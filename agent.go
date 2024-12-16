package ants

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	semverRegex = regexp.MustCompile(`.*(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*).*`)
	hexRegex    = regexp.MustCompile(`^[a-fA-F0-9]+$`)
)

type agentVersionInfo struct {
	full  string
	typ   string
	major int
	minor int
	patch int
	hash  string
}

func (avi *agentVersionInfo) Semver() [3]int {
	return [3]int{avi.major, avi.minor, avi.patch}
}

func parseAgentVersion(av string) agentVersionInfo {
	avi := agentVersionInfo{
		full: av,
	}

	switch {
	case av == "":
		return avi
	case av == "celestia-celestia":
		avi.typ = "celestia-celestia"
		return avi
	case strings.HasPrefix(av, "celestia-node/celestia/"):
		// fallthrough
	default:
		avi.typ = "other"
		return avi
	}

	parts := strings.Split(av, "/")
	if len(parts) > 2 {
		switch parts[2] {
		case "bridge", "full", "light":
			avi.typ = parts[2]
		default:
			avi.typ = "other"
		}
	}

	if len(parts) > 3 {
		matches := semverRegex.FindStringSubmatch(parts[3])
		if matches != nil {
			for i, name := range semverRegex.SubexpNames() {
				switch name {
				case "major":
					avi.major, _ = strconv.Atoi(matches[i])
				case "minor":
					avi.minor, _ = strconv.Atoi(matches[i])
				case "patch":
					avi.patch, _ = strconv.Atoi(matches[i])
				}
			}
		}
	}

	if len(parts) > 4 && hexRegex.MatchString(parts[4]) {
		avi.hash = parts[4]
	}

	return avi
}
