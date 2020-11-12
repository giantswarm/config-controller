package github

import (
	"reflect"
	"strconv"
	"testing"
)

func Test_getLatestTag(t *testing.T) {
	testCases := []struct {
		name        string
		inputTags   []string
		expectedTag string
	}{
		{
			name: "case 0: nil",
			inputTags: []string{
				"",
				"non-version-tag",
			},
			expectedTag: "",
		},
		{
			name: "case 1: empty",
			inputTags: []string{
				"",
				"non-version-tag",
			},
			expectedTag: "",
		},
		{
			name: "case 2: major",
			inputTags: []string{
				"v2.0.0",
				"v1.2.3",
				"non-version-tag",
			},
			expectedTag: "v2.0.0",
		},
		{
			name: "case 3: minor",
			inputTags: []string{
				"v2.5.9",
				"non-version-tag",
				"v2.10.3",
				"v2.2.1",
			},
			expectedTag: "v2.10.3",
		},
		{
			name: "case 4: patch",
			inputTags: []string{
				"non-version-tag",
				"v2.10.300",
				"v1.5.500",
				"v2.10.370",
				"v2.10.372",
				"v2.10.301",
				"v2.10.360",
			},
			expectedTag: "v2.10.372",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			tag := getLatestTag(tc.inputTags)

			if !reflect.DeepEqual(tag, tc.expectedTag) {
				t.Fatalf("tag = %q, want %q", tag, tc.expectedTag)
			}
		})
	}
}
