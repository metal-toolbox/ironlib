package utils

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// old and new versions are semver
func Test_VersioNewIsNewer(t *testing.T) {

	b, err := VersionIsNewer("3.4", "3.3")
	assert.Equal(t, nil, err)
	assert.Equal(t, true, b)

	b, err = VersionIsNewer("3.3", "3.4")
	assert.Equal(t, nil, err)
	assert.Equal(t, false, b)

}

// Older and newer versions are non semver, equal
func Test_VersionBothNotSemverEqual(t *testing.T) {
	b, err := VersionIsNewer("ABC", "ABC")
	assert.Equal(t, nil, err)
	assert.Equal(t, false, b)
}

// Older and newer versions are non semver, un-equal
func Test_VersionBothNotSemverUnEqual(t *testing.T) {
	b, err := VersionIsNewer("ABC", "EFG")
	assert.Equal(t, nil, err)
	assert.Equal(t, true, b)
}

// Older version is semver, but newer version specified as non semver
func Test_VersionNewIsNotSemver(t *testing.T) {
	b, err := VersionIsNewer("ABC", "3.3")
	assert.Equal(t, ErrVersionStrExpectedSemver, errors.Cause(err))
	assert.Equal(t, false, b)
}

// old version is non semver, new version is semver
func Test_VersionOldIsNotSemver(t *testing.T) {
	b, err := VersionIsNewer("3.3", "ABC")
	assert.Equal(t, nil, err)
	assert.Equal(t, true, b)
}

func Test_ValidateSHA1Checksum(t *testing.T) {
	err := ValidateSHA1Checksum("test_data/samplefile", "test_data/samplefile.sha1")
	assert.Equal(t, nil, err)
}
