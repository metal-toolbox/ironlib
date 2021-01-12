package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_VersionIsNewer(t *testing.T) {

	b, err := VersionIsNewer("3.4", "3.3")
	assert.Equal(t, nil, err)
	assert.Equal(t, true, b)

	b, err = VersionIsNewer("3.3", "3.4")
	assert.Equal(t, nil, err)
	assert.Equal(t, false, b)

}

func Test_ValidateSHA1Checksum(t *testing.T) {
	err := ValidateSHA1Checksum("test_data/samplefile", "test_data/samplefile.sha1")
	assert.Equal(t, nil, err)
}
