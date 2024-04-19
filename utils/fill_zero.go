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

func (z *FillZero) WipeDisk(ctx context.Context, path string) error {
	log.Println("Starting zero-fill of", path)
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
	log.Printf("%s | Size: %dB\n", path, partitionSize)
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	var bytesSinceLastPrint int64
	var totalBytesWritten int64
	buffer := make([]byte, 4096)
	start := time.Now()
	for bytesRemaining := partitionSize; bytesRemaining > 0; {
		l := min(int64(len(buffer)), bytesRemaining)
		bytesWritten, err := file.Write(buffer[:l])
		if err != nil {
			return err
		}
		totalBytesWritten += int64(bytesWritten)
		bytesSinceLastPrint += int64(bytesWritten)
		bytesRemaining -= int64(bytesWritten)
		// Print progress report every 10 seconds and when done
		if bytesRemaining == 0 || time.Since(start) >= 10*time.Second {
			printProgress(totalBytesWritten, partitionSize, &start, &bytesSinceLastPrint, bytesRemaining, path)
		}
	}
	err = file.Sync()
	if err != nil {
		return err
	}
	return nil
}

func printProgress(totalBytesWritten, partitionSize int64, start *time.Time, bytesSinceLastPrint *int64, bytesRemaining int64, path string) {
	// Calculate progress and ETA
	progress := float64(totalBytesWritten) / float64(partitionSize) * 100
	elapsed := time.Since(*start).Seconds()
	speed := float64(*bytesSinceLastPrint) / elapsed                                  // Speed in bytes per second
	remainingSeconds := (float64(partitionSize) - float64(totalBytesWritten)) / speed // Remaining time in seconds
	remainingHours := float64(remainingSeconds / 3600)
	mbPerSecond := speed / (1024 * 1024)
	log.Printf("%s | Progress: %.2f%% | Speed: %.2f MB/s | Estimated time left: %.2f hour(s)\n", path, progress, mbPerSecond, remainingHours)
	*start = time.Now()
	*bytesSinceLastPrint = 0
}

// We are in go 1.19 min not available yet
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
