package utils

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	EnvZeroWipeUtility = "IRONLIB_UTIL_WIPE_ZERO"
)

type ZeroWipe struct {
}

var (
	ErrWipeDisk = errors.New("failed to wipe disk")
)

// Return a new zerowipe executor
func NewZeroWipeCmd(trace bool) *ZeroWipe {
	return &ZeroWipe{}
}

func (z *ZeroWipe) Wipe(ctx context.Context, logicalName string) error {
	log.Println("Start wiping with zeros...")
	partitionPath := logicalName // example /dev/sdb

	// Buffer size (in bytes)
	bufferSize := 4096 // to use 1MB = 1024 * 1024

	// Write open
	file, err := os.OpenFile(partitionPath, os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open :"+logicalName, err)
		return err
	}
	defer file.Close()

	// Get disk or partition size
	partitionSize, err := file.Seek(0, 2)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(partitionSize)

	// Create a buffer fill with zeroes
	buffer := make([]byte, bufferSize)

	var bytesWritten int64
	start := time.Now()

	// rewind cassete tape ;=)
	file.Seek(0, 0)

	// Writting zeroes loop
	for bytesWritten < partitionSize {
		n, err := file.Write(buffer)
		if err != nil {
			if strings.Contains(err.Error(), "no space left on device") { //syscall.ENOSPC
				// If partitionSize % bufferSize is not 0 the last write loop will not write all the buffer size
				log.Println("we have reached the end of the device.")
			} else {
				// Other errors
				log.Println("failed to write to disk:", err)
			}
			return err
		}
		bytesWritten += int64(n)

		// Print progress each 10 seconds
		if time.Since(start) >= 10*time.Second {
			log.Printf("Progress: %.2f%%\n", float64(bytesWritten)/float64(partitionSize)*100)
			start = time.Now()
		}
	}

	log.Println("Device has been completely overwritten with zeros")

	return nil
}

func (z *ZeroWipe) WipeDisk(ctx context.Context, logicalName string) error {
	return z.Wipe(ctx, logicalName)
}
