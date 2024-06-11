package utils

import (
	"crypto/rand"
	"io"
	"os"
	"testing"

	"github.com/bmc-toolbox/common"
	"github.com/stretchr/testify/assert"
)

func createTestDrive(t *testing.T) *common.Drive {
	// Create a temporary directory
	// go will clean up the whole directory tree when the test is done
	dir := t.TempDir()

	f, err := os.Create(dir + "/test-file")
	assert.NoError(t, err)
	assert.NoError(t, f.Close())
	return &common.Drive{Common: common.Common{LogicalName: f.Name()}}
}

func Test_ApplyWatermarks(t *testing.T) {
	t.Run("EmptyFile", func(t *testing.T) {
		tempFile := createTestDrive(t)

		checker, err := ApplyWatermarks(tempFile)
		assert.Nil(t, checker)
		assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
	})

	t.Run("NotEnoughSpace", func(t *testing.T) {
		tempFile := createTestDrive(t)

		// Create a 1KB empty file, no room for all watermarks
		assert.NoError(t, os.Truncate(tempFile.LogicalName, 1*1024))

		checker, err := ApplyWatermarks(tempFile)
		assert.Nil(t, checker)
		assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
	})

	t.Run("WipeFailed", func(t *testing.T) {
		tempFile := createTestDrive(t)

		// Write the file full of random data
		randomData := make([]byte, 15*1024*1024)
		_, err := rand.Read(randomData)
		assert.NoError(t, err)
		assert.NoError(t, os.WriteFile(tempFile.LogicalName, randomData, 0o600))

		// Apply watermarks and expect no error
		checker, err := ApplyWatermarks(tempFile)
		assert.NoError(t, err)

		assert.ErrorIs(t, checker(), ErrIneffectiveWipe)
	})

	t.Run("WipeSucceeded", func(t *testing.T) {
		tempFile := createTestDrive(t)

		// Write the file full of random data
		randomData := make([]byte, 15*1024*1024)
		_, err := rand.Read(randomData)
		assert.NoError(t, err)
		assert.NoError(t, os.WriteFile(tempFile.LogicalName, randomData, 0o600))

		// Apply watermarks and expect no error
		checker, err := ApplyWatermarks(tempFile)
		assert.NoError(t, err)

		// simulate wipe
		assert.NoError(t, os.Truncate(tempFile.LogicalName, 0))
		assert.NoError(t, os.Truncate(tempFile.LogicalName, 15*1024*1024))

		assert.NoError(t, checker())
	})
}
