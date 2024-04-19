package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

const (
	EnvFillZeroUtility = "IRONLIB_UTIL_FILL_ZERO"
)

type FillZero struct {
}

// Return a new zerowipe executor
func NewFillZeroCmd(trace bool) *FillZero {
	return &FillZero{}
}

//noline:gomnd // the magic numbers here are fine
func (z *FillZero) WipeDisk(ctx context.Context, path string) error {
	fmt.Println("Starting zero-fill of", path)
	// Write open
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		log.Println(fmt.Errorf("%w", err))
		return err
	}
	defer file.Close()
	// Get disk or partition size
	partitionSize, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Printf("%s | Size: %dB\n", path, partitionSize)
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	var bytesSinceLastPrint int
	var totalBytesWritten int
	buffer := make([]byte, 4096)
	start := time.Now()
	for bytesRemaining := int(partitionSize); bytesRemaining > 0; {
		l := min(len(buffer), bytesRemaining)
		bytesWritten, err := file.Write(buffer[:l])
		if err != nil {
			return err
		}
		totalBytesWritten += bytesWritten
		bytesSinceLastPrint += bytesWritten
		bytesRemaining -= bytesWritten
		// Print progress report every 10 seconds and when done
		if bytesRemaining == 0 || time.Since(start) >= 10*time.Second {
			progress := float64(totalBytesWritten) / float64(partitionSize) * 100
			elapsed := time.Since(start)
			rate := int(int64(bytesSinceLastPrint) / elapsed.Nanoseconds())
			remaining := time.Duration(bytesRemaining / rate)
			mbPerSecond := float64(time.Duration(rate)*time.Second) / (1 * 1024 * 1024)
			fmt.Printf("%s | Progress: %6.2f%% | Speed: %.2f MB/s | Estimated time left: %5.2fh\n",
				path, progress, mbPerSecond, remaining.Hours())
			start = time.Now()
			bytesSinceLastPrint = 0
		}
	}
	file.Sync()
	return nil
}

// We are in go 1.19 min not available yet
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
