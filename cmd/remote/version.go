package remote

import (
	"fmt"
	"regexp"
	"strconv"
)

// defaultVersion presents a version in format:
// Major.Minor.Patch-Tag
type defaultVersion struct {
	Major int
	Minor int
	Patch int
	Tag   string
}

type defaultVersionList []defaultVersion

func toInt(str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}

	return val
}

func ParseVersion(versionAsString string, version *defaultVersion) error {
	t := regexp.MustCompile(`([0-9]+)\.?([0-9]+)?\.?([0-9]+)?([-_]{1}([a-zA-Z0-9]+))?`)
	if t.MatchString(versionAsString) {
		fields := t.FindStringSubmatch(versionAsString)
		*version = defaultVersion{
			Major: toInt(fields[1]),
			Minor: toInt(fields[2]),
			Patch: toInt(fields[3]),
			Tag:   fields[5],
		}
		return nil
	}

	return fmt.Errorf("version does not match with pattern (major.minor.path-tag)")
}

func (version defaultVersion) String() string {
	if version.Tag != "" {
		return fmt.Sprintf("%d.%d.%d-%s", version.Major, version.Minor, version.Patch, version.Tag)
	}

	return fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)
}

func Less(l defaultVersion, r defaultVersion) bool {
	lseg := [3]int{l.Major, l.Minor, l.Patch}
	rseg := [3]int{r.Major, r.Minor, r.Patch}
	for i := 0; i < 3; i++ {
		if lseg[i] != rseg[i] {
			return lseg[i] < rseg[i]
		}
	}

	// If segments are equal, then compare the prerelease info
	return l.Tag < r.Tag
}

func IsVersionSmaller(v1 string, v2 string) bool {
	var l defaultVersion
	var r defaultVersion

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

func (lst defaultVersionList) Len() int {
	return len(lst)
}

func (lst defaultVersionList) Less(i, j int) bool {
	return Less(lst[i], lst[j])
}

func (lst defaultVersionList) Swap(i, j int) {
	lst[i], lst[j] = lst[j], lst[i]
}
