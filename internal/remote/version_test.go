package remote

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToString(t *testing.T) {
	var version defaultVersion
	err := ParseVersion("1.2.3-test", &version)
	if err != nil {
		t.Fail()
	}

	if version.String() != "1.2.3-test" {
		t.Fail()
	}
}

func TestParseVersion(t *testing.T) {
	var version defaultVersion
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

	var version defaultVersion
	for _, verAsString := range versions {
		err := ParseVersion(verAsString, &version)
		if err != nil {
			t.Fail()
		}
	}
}

func TestSorter(t *testing.T) {
	values := []string{
		"2.0.2",
		"2.0.1",
		"2",
		"1.2",
		"1",
		"1.2.3",
		"1.0.9",
	}

	var versions []defaultVersion
	for _, val := range values {
		var version defaultVersion
		_ = ParseVersion(val, &version)
		versions = append(versions, version)
	}

	sort.Sort(defaultVersionList(versions))

	expected_versions := []defaultVersion{
		{1, 0, 0, ""},
		{1, 0, 9, ""},
		{1, 2, 0, ""},
		{1, 2, 3, ""},
		{2, 0, 0, ""},
		{2, 0, 1, ""},
		{2, 0, 2, ""},
	}

	for i := 0; i < len(expected_versions); i++ {
		assert.Equal(t, expected_versions[i], versions[i], fmt.Sprintf("versions[%d] should be %v", i, expected_versions[i]))
	}
}

func TestLess(t *testing.T) {
	test_cases := []struct {
		l, r    defaultVersion
		compare int
	}{
		{defaultVersion{1, 0, 0, ""}, defaultVersion{1, 0, 1, ""}, -1},
		{defaultVersion{1, 0, 0, ""}, defaultVersion{1, 0, 0, ""}, 0},
		{defaultVersion{1, 0, 2, ""}, defaultVersion{1, 1, 0, ""}, -1},
		{defaultVersion{1, 1, 0, ""}, defaultVersion{1, 0, 1, ""}, 1},
		{defaultVersion{1, 0, 0, "rc1"}, defaultVersion{1, 0, 0, "rc2"}, -1},
		{defaultVersion{1, 0, 0, "rc2"}, defaultVersion{1, 0, 0, "rc2"}, 0},
	}
	for _, tc := range test_cases {
		assert.Equal(t, tc.compare < 0, Less(tc.l, tc.r), fmt.Sprintf("Less(%v, %v) should be %v", tc.l, tc.r, tc.compare < 0))
		assert.Equal(t, tc.compare > 0, Less(tc.r, tc.l), fmt.Sprintf("Less(%v, %v) should be %v", tc.r, tc.l, tc.compare > 0))
	}
}
