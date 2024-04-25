package utils

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"os"
	"slices"

	"github.com/pkg/errors"
)

const (
	bufferSize    = 512
	numWatermarks = 10
)

type watermark struct {
	position int64
	data     []byte
}

func ApplyWatermarks(logicalName string) (func() error, error) {
	// Write open
	file, err := os.OpenFile(logicalName, os.O_WRONLY, 0)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get disk or partition size
	fileSize, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	// Write watermarks on random locations
	watermarks := writeWatermarks(file, 0, fileSize, numWatermarks)
	if len(watermarks) != numWatermarks {
		return nil, err
	}

	checker := func() error {
		file, err := os.OpenFile(logicalName, os.O_RDONLY, 0)
		if err != nil {
			return err
		}
		defer file.Close()

		for _, watermark := range watermarks {
			_, err = file.Seek(watermark.position, io.SeekStart)
			if err != nil {
				return err
			}
			// Read the watermark written to the position
			currentValue := make([]byte, bufferSize)
			_, err = io.ReadFull(file, currentValue)
			if err != nil {
				return err
			}
			// Check if the watermark is still in the disk
			if slices.Equal(currentValue, watermark.data) {
				ErrorExistingWatermark := errors.New("Error existing watermark in the file")
				return fmt.Errorf("%s | %w", logicalName, ErrorExistingWatermark)
			}
		}
		return nil
	}
	return checker, nil
}

func writeWatermarks(file *os.File, a, b int64, count int) []watermark {
	if count == 1 {
		data := make([]byte, bufferSize)
		_, err := rand.Read(data)
		if err != nil {
			return nil
		}
		offset, err := rand.Int(rand.Reader, big.NewInt(b-a-bufferSize))
		if err != nil {
			return nil
		}
		randomPosition := int64(offset.Uint64()) + a
		_, err = file.Seek(randomPosition, io.SeekStart)
		if err != nil {
			return nil
		}
		_, err = file.Write(data)
		if err != nil {
			return nil
		}

		w := watermark{
			position: randomPosition,
			data:     data,
		}
		return []watermark{w}
	}
	// Divide the intervals into two equal parts (approximately)
	mid := (a + b) / 2

	// Return recursively the call of function with the two remaining intervals
	leftCount := count / 2
	rightCount := count - leftCount

	return append(writeWatermarks(file, a, mid-1, leftCount), writeWatermarks(file, mid, b, rightCount)...)
}
