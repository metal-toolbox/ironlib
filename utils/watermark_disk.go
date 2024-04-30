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

var ErrIneffectiveWipe = errors.New("found left over data after wiping disk")

type watermark struct {
	position int64
	data     []byte
}

// ApplyWatermarks applies watermarks to the specified disk.
// It returns a function that checks if the applied watermarks still exist on the file.
// It relies on the writeWatermarks function to uniformly write watermarks across the disk.
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

	if fileSize == 0 {
		return nil, errors.New("No space for watermarking")
	}

	// Write watermarks on random locations
	watermarks, err := writeWatermarks(file, fileSize, numWatermarks)
	if err != nil {
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
				return fmt.Errorf("verify wipe %s@%d: %w", logicalName, watermark.position, ErrIneffectiveWipe)
			}
		}
		return nil
	}
	return checker, nil
}

// writeWatermarks creates random watermarks and writes them uniformly into a given file.
func writeWatermarks(file *os.File, fileSize, count int64) ([]watermark, error) {
	origin := int64(0)
	intervalSize := fileSize / count
	watermarks := make([]watermark, count)
	for i := 0; i < numWatermarks; i++ {
		data := make([]byte, bufferSize)
		_, err := rand.Read(data)
		if err != nil {
			return nil, err
		}
		offset, err := rand.Int(rand.Reader, big.NewInt(intervalSize))
		if err != nil {
			return nil, err
		}
		randomPosition := int64(offset.Uint64()) + origin - bufferSize
		_, err = file.Seek(randomPosition, io.SeekStart)
		if err != nil {
			return nil, err
		}
		_, err = file.Write(data)
		if err != nil {
			return nil, err
		}
		watermarks[i].position = randomPosition
		watermarks[i].data = data
		origin += intervalSize
	}
	return watermarks, nil
}
