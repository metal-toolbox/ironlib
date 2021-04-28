package utils

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionIsNewer(t *testing.T) {

	type testCases struct {
		testName    string
		newVersion  string
		oldVersion  string
		isNewer     bool
		expectError bool
		err         error
	}

	var cases = []testCases{
		{"newVersion, newer", "3.4", "3.3", true, false, nil},
		{"oldVersion, newer", "3.3", "3.4", false, false, nil},
		{"versions, equal", "3.3", "3.3", false, false, nil},
		{"non-semver, equal", "ABC", "ABC", false, false, nil},
		{"new non-semver, unequal", "ABC", "3.3", false, true, ErrVersionStrExpectedSemver},
		{"old non-semver, unequal", "3.3", "ABC", true, false, nil},
	}

	for _, tt := range cases {
		b, err := VersionIsNewer(tt.newVersion, tt.oldVersion)
		if tt.expectError {
			assert.Equal(t, tt.err, errors.Cause(err), tt.testName)
		} else {
			require.NoError(t, err, tt.testName)
			assert.Equal(t, tt.isNewer, b, tt.testName)
		}
	}
}

func Test_ValidateSHA1Checksum(t *testing.T) {
	err := ValidateSHA1Checksum("test_data/samplefile", "test_data/samplefile.sha1")
	assert.Equal(t, nil, err)
}
