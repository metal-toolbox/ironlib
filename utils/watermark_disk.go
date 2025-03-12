package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"slices"
	"time"

	common "github.com/metal-toolbox/bmc-common"
)

var ErrIneffectiveWipe = errors.New("found left over data after wiping disk")

type watermark struct {
	position int64
	data     []byte
}

// ApplyWatermarks applies random watermarks randomly through out the specified device/file.
// It returns a function that checks if the applied watermarks still exists on the device/file.
func ApplyWatermarks(drive *common.Drive) (func() error, error) {
	// Write open
	file, err := os.OpenFile(drive.LogicalName, os.O_WRONLY|os.O_SYNC, 0)
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
		// The delay gives the controller time to release I/O blocking, which could otherwise cause the verification process to fail due to incomplete or pending I/O operations.
		time.Sleep(500 * time.Millisecond)
		checkFile, checkErr := os.OpenFile(drive.LogicalName, os.O_RDONLY, 0)
		if checkErr != nil {
			return checkErr
		}
		defer checkFile.Close()

		for i, watermark := range watermarks {
			_, checkErr = checkFile.Seek(watermark.position, io.SeekStart)
			if checkErr != nil {
				return fmt.Errorf("watermark verification, %s@%d(mark=%d), seek: %w", drive.LogicalName, watermark.position, i, checkErr)
			}

			// Read the watermark written to the position
			currentValue := make([]byte, watermarkSize)
			_, checkErr = io.ReadFull(checkFile, currentValue)
			if checkErr != nil {
				return fmt.Errorf("read watermark %s@%d(mark=%d): %w", drive.LogicalName, watermark.position, i, checkErr)
			}

			// Check if the watermark is still in the disk
			if slices.Equal(currentValue, watermark.data) {
				return fmt.Errorf("verify wipe %s@%d(mark=%d): %w", drive.LogicalName, watermark.position, i, ErrIneffectiveWipe)
			}
		}
		return nil
	}
	// We introduce a 500-millisecond delay to give the OS enough time to properly flush the disk buffers to disk.
	// While this delay helps ensure that the data is written, it is not an ideal solution, and further investigation is needed to find more efficient synchronization mechanisms.
	err = file.Sync()
	if err != nil {
		return nil, err
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}
	time.Sleep(500 * time.Millisecond)
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
	for chunkStart, i := int64(0), int64(0); i < watermarksCount; chunkStart, i = chunkStart+chunkSize, i+1 {
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
