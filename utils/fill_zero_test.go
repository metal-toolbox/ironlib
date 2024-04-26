package utils

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"
)

func Test_NewFillZeroCmd(t *testing.T) {
	// Test if NewFillZeroCmd returns a non-nil pointer
	zw := NewFillZeroCmd(false)
	if zw == nil {
		t.Error("Expected non-nil pointer, got nil")
	}
}

func Test_WipeDisk(t *testing.T) {
	for _, size := range []int{4095, 4096, 4097, 8192} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			// Create a temporary file for testing
			tmpfile, err := os.CreateTemp("", "example")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name()) // clean up

			// Write some content to the temporary file
			expectedSize := int64(4096)
			if _, err = tmpfile.Write(make([]byte, expectedSize)); err != nil {
				t.Fatal(err)
			}

			// Simulate a context
			ctx := context.Background()

			// Create a FillZero instance
			zw := &FillZero{}

			// Test Fill function
			err = zw.WipeDisk(ctx, logrus.New(), tmpfile.Name())
			if err != nil {
				t.Errorf("Fill returned an error: %v", err)
			}

			// Check if the file size remains the same after overwrite
			fileInfo, err := os.Stat(tmpfile.Name())
			if err != nil {
				t.Fatal(err)
			}

			if size := fileInfo.Size(); size != expectedSize {
				t.Errorf("Expected file size to remain %d after overwrite, got %d", expectedSize, size)
			}
		})
	}
}
