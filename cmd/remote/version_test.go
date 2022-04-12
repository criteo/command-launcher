package remote

import (
	"sort"
	"testing"
)

func TestToString(t *testing.T) {
	var version cdtVersion
	err := ParseVersion("1.2.3-test", &version)
	if err != nil {
		t.Fail()
	}

	if version.String() != "1.2.3-test" {
		t.Fail()
	}
}

func TestParseVersion(t *testing.T) {
	var version cdtVersion
	err := ParseVersion("1.2.3-test", &version)
	if err != nil {
		t.Fail()
	}

	if version.Major != 1 {
		t.Errorf("Unknow major %d", version.Major)
	}

	if version.Minor != 2 {
		t.Errorf("Unknow minor %d", version.Minor)
	}

	if version.Patch != 3 {
		t.Errorf("Unknow patch %d", version.Patch)
	}

	if version.Tag != "test" {
		t.Errorf("Unknow Tag %s", version.Tag)
	}
}

func TestParseVersions(t *testing.T) {
	versions := []string{
		"1",
		"1.2",
		"1.2.3",
		"1-tag",
		"1.2-tag",
		"1.2.3-tag",
		"1.2.3_tag",
	}

	var version cdtVersion
	for _, verAsString := range versions {
		err := ParseVersion(verAsString, &version)
		if err != nil {
			t.Fail()
		}
	}
}

func TestSorter(t *testing.T) {
	values := []string{
		"2",
		"1.2",
		"1",
		"1.2.3",
	}

	var versions []cdtVersion
	for _, val := range values {
		var version cdtVersion
		_ = ParseVersion(val, &version)
		versions = append(versions, version)
	}

	sort.Sort(cdtVersionList(versions))
	if versions[0].Major != 1 {
		t.Log(versions)
	}

	if versions[1].Minor != 2 {
		t.Log(versions)
	}

	if versions[2].Patch != 3 {
		t.Log(versions)
	}

	if versions[3].Major != 2 {
		t.Log(versions)
	}
}
