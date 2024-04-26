package utils

import (
	"crypto/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ApplyWatermarks(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Close the file since we'll be reopening it in ApplyWatermarks
	tempFile.Close()

	t.Run("NegativeTest", func(t *testing.T) {
		// Create a ~15KB empty file, no room for all watermarks
		err := os.WriteFile(tempFile.Name(), make([]byte, 1*1024), 0o600)
		if err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}
		// Create a 1KB empty file, no room for all watermarks
		assert.NoError(t, os.Truncate(tempFile.Name(), 1*1024))
		// Apply watermarks and expect an error
		checker, _ := ApplyWatermarks(tempFile.Name())
		assert.Nil(t, checker)
	})

	t.Run("EmptyFile", func(t *testing.T) {
		// Wipe the file
		assert.NoError(t, os.Truncate(tempFile.Name(), 0))

		// Apply watermarks and expect no error
		checker, err := ApplyWatermarks(tempFile.Name())
		assert.Error(t, err, "No space for watermarking")
		assert.Nil(t, checker)
	})
	t.Run("PositiveTestWithRandomDataAndWipe", func(t *testing.T) {
		// Write the file full of random data
		randomData := make([]byte, 15*1024*1024)
		_, err := rand.Read(randomData)
		if err != nil {
			t.Fatalf("Failed to generate random data: %v", err)
		}
		err = os.WriteFile(tempFile.Name(), randomData, 0o600)
		if err != nil {
			t.Fatalf("Failed to write random data to file: %v", err)
		}

		// Apply watermarks and expect no error
		checker, err := ApplyWatermarks(tempFile.Name())
		if err != nil {
			t.Fatalf("Error applying watermarks: %v", err)
		}
		// simulate wipe
		assert.NoError(t, os.Truncate(tempFile.Name(), 0))
		assert.NoError(t, os.Truncate(tempFile.Name(), 15*1024*1024))

		err = checker()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})
}
