package remote

import (
	"fmt"
	"regexp"
	"strconv"
)

// cdtVersion presents a version in format:
// Major.Minor.Patch-Tag
type cdtVersion struct {
	Major int
	Minor int
	Patch int
	Tag   string
}

type cdtVersionList []cdtVersion

func toInt(str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}

	return val
}

func ParseVersion(versionAsString string, version *cdtVersion) error {
	t := regexp.MustCompile(`([0-9]+)\.?([0-9]+)?\.?([0-9]+)?([-_]{1}([a-zA-Z0-9]+))?`)
	if t.MatchString(versionAsString) {
		fields := t.FindStringSubmatch(versionAsString)
		*version = cdtVersion{
			Major: toInt(fields[1]),
			Minor: toInt(fields[2]),
			Patch: toInt(fields[3]),
			Tag:   fields[5],
		}
		return nil
	}

	return fmt.Errorf("version does not match with pattern (major.minor.path-tag)")
}

func (version cdtVersion) String() string {
	if version.Tag != "" {
		return fmt.Sprintf("%d.%d.%d-%s", version.Major, version.Minor, version.Patch, version.Tag)
	}

	return fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)
}

func Less(l cdtVersion, r cdtVersion) bool {
	if l.Major < r.Major {
		return true
	} else if l.Minor < r.Minor {
		return true
	} else if l.Patch < r.Patch {
		return true
	}

	return false
}

func IsVersionSmaller(v1 string, v2 string) bool {
	var l cdtVersion
	var r cdtVersion

	err := ParseVersion(v1, &l)
	if err != nil { // wrong format version is considered smaller
		return true
	}
	err = ParseVersion(v2, &r)
	if err != nil { // wrong fromat version is considered smaller
		return false
	}

	return Less(l, r)
}

func (lst cdtVersionList) Len() int {
	return len(lst)
}

func (lst cdtVersionList) Less(i, j int) bool {
	return Less(lst[i], lst[j])
}

func (lst cdtVersionList) Swap(i, j int) {
	lst[i], lst[j] = lst[j], lst[i]
}
