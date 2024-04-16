package utils

import (
	"context"
	"log"
	"os"
	"strings"
	"time"
)

const (
	EnvFillZeroUtility = "IRONLIB_UTIL_FILL_ZERO"
	FileMode644        = 0o644 // Owner readwrite, group and others readonly
	OneMBinBytes       = 1024 * 1024
	SecondsPerHour     = 3600
	TotalPercentage    = 100
)

type FillZero struct {
}

// Return a new zerowipe executor
func NewFillZeroCmd(trace bool) *FillZero {
	return &FillZero{}
}

func (z *FillZero) Fill(ctx context.Context, logicalName string) error {
	log.Printf("Start overwriting with zeros %s", logicalName)
	partitionPath := logicalName // example /dev/sdb

	// Buffer size (in bytes)
	bufferSize := 4096 // to use 1MB = 1024 * 1024

	// Write open
	file, err := os.OpenFile(partitionPath, os.O_WRONLY, FileMode644)
	if err != nil {
		log.Println("Failed to open :"+logicalName, err)
		return err
	}
	defer file.Close()

	// Get disk or partition size
	whence := 2 // 2 means relative to the end, check Seek function
	partitionSize, err := file.Seek(0, whence)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("%s | Size in bytes: %d \n", partitionPath, partitionSize)

	// Create a buffer fill with zeroes
	buffer := make([]byte, bufferSize)

	var totalBytesWritten int64
	var bytesSinceLastPrint int64
	start := time.Now()

	// rewind cassette tape ;=)
	whence = 0 // 0 means relative to the origin of the file, check Seek function
	_, err = file.Seek(0, whence)
	if err != nil {
		log.Println(err)
		return err
	}

	// Writing zeroes loop
	bytesSinceLastPrint = 0
	for totalBytesWritten < partitionSize {
		bytesWritten, err := file.Write(buffer)
		if err != nil {
			if strings.Contains(err.Error(), "no space left on device") { // syscall.ENOSPC
				// Check if the last write loop didn't write all the buffer size, the reason is partitionSize % bufferSize is not 0
				if totalBytesWritten+int64(bytesWritten) == partitionSize {
					log.Println("we have reached the end of the devices.")
					return nil
				} else {
					log.Println("we have a problem with the write to disk: ", err)
				}
			} else {
				// Other errors
				log.Println("failed to write to disk: ", err)
			}
			return err
		}
		totalBytesWritten += int64(bytesWritten)
		bytesSinceLastPrint += int64(bytesWritten)

		// Print progress each 10 seconds
		if time.Since(start) >= 10*time.Second {
			// Calculate progress and ETA
			progress := float64(totalBytesWritten) / float64(partitionSize) * TotalPercentage
			elapsed := time.Since(start).Seconds()
			speed := float64(bytesSinceLastPrint) / elapsed                                   // Speed in bytes per second
			remainingSeconds := (float64(partitionSize) - float64(totalBytesWritten)) / speed // Remaining time in seconds
			remainingHours := float64(remainingSeconds / SecondsPerHour)
			mbPerSecond := speed / OneMBinBytes
			log.Printf("%s | Progress: %.2f%% | Speed: %.2f MB/s | Estimated time left: %.2f hour(s)\n", partitionPath, progress, mbPerSecond, remainingHours)
			start = time.Now()
			bytesSinceLastPrint = 0
		}
	}

	log.Printf("%s has been completely overwritten with zeros", partitionPath)

	return nil
}

func (z *FillZero) WipeDisk(ctx context.Context, logicalName string) error {
	return z.Fill(ctx, logicalName)
}
