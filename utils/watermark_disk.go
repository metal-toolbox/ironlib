package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"slices"
)

var ErrIneffectiveWipe = errors.New("found left over data after wiping disk")

type watermark struct {
	position int64
	data     []byte
}

// ApplyWatermarks applies random watermarks randomly through out the specified device/file.
// It returns a function that checks if the applied watermarks still exists on the device/file.
func ApplyWatermarks(logicalName string) (func() error, error) {
	// Write open
	file, err := os.OpenFile(logicalName, os.O_WRONLY, 0)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Write watermarks on random locations
	watermarkSize := int64(512)
	watermarks, err := writeWatermarks(file, 10, watermarkSize)
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
			currentValue := make([]byte, watermarkSize)
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

// writeWatermarks creates random watermarks and writes them randomlyish throughout the given file.
//
// It calculates a chunksize by dividing the file size by _watermarksCount_.
// The disk is sectioned off into these chunks.
// We will generate a random byte slice of _watermarksSize_ that will be written into a chunk.
// The position within the chunk is randomly generated such that the write does not spill into the next chunk.
// Both the random bytes and sub-chunk start position is independently generated for each chunk.
func writeWatermarks(file *os.File, watermarksCount, watermarksSize int64) ([]watermark, error) {
	// Get disk or partition watermarksSize
	fileSize, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	if fileSize <= watermarksCount*watermarksSize {
		return nil, fmt.Errorf("no space for watermarking: %w", io.ErrUnexpectedEOF)
	}
	chunkSize := fileSize / watermarksCount

	watermarks := make([]watermark, watermarksCount)
	for chunkStart, i := int64(0), 0; chunkStart < fileSize; chunkStart, i = chunkStart+chunkSize, i+1 {
		data := make([]byte, watermarksSize)
		_, err := rand.Read(data)
		if err != nil {
			return nil, err
		}

		offset, err := rand.Int(rand.Reader, big.NewInt(chunkSize-watermarksSize))
		if err != nil {
			return nil, err
		}

		randomPosition := chunkStart + offset.Int64()
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
	}

	return watermarks, nil
}
