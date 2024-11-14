package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	common "github.com/metal-toolbox/bmc-common"
	"github.com/sirupsen/logrus"
)

type FillZero struct {
	Quiet bool
}

// Return a new zerowipe executor
func NewFillZeroCmd(trace bool) *FillZero {
	z := FillZero{}
	if !trace {
		z.SetQuiet()
	}
	return &z
}

func (z *FillZero) WipeDrive(ctx context.Context, logger *logrus.Logger, drive *common.Drive) error {
	log := logger.WithField("drive", drive.LogicalName).WithField("method", "zero-fill")
	log.Debug("wiping")

	verify, err := ApplyWatermarks(drive)
	if err != nil {
		return err
	}

	// Write open
	file, err := os.OpenFile(drive.LogicalName, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get disk or partition size
	partitionSize, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	log.WithField("size", fmt.Sprintf("%dB", partitionSize)).Debug("disk info detected")
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	var bytesSinceLastPrint int64
	var totalBytesWritten int64
	buffer := make([]byte, 4096)
	start := time.Now()
	for bytesRemaining := partitionSize; bytesRemaining > 0; {
		// Check if the context has been canceled
		select {
		case <-ctx.Done():
			log.Debug("stopping")
			return ctx.Err()
		default:
			l := min(int64(len(buffer)), bytesRemaining)
			bytesWritten, writeError := file.Write(buffer[:l])
			if writeError != nil {
				return writeError
			}

			totalBytesWritten += int64(bytesWritten)
			bytesSinceLastPrint += int64(bytesWritten)
			bytesRemaining -= int64(bytesWritten)
			// Print progress report every 10 seconds and when done
			if bytesRemaining == 0 || time.Since(start) >= 10*time.Second {
				printProgress(log, totalBytesWritten, partitionSize, start, bytesSinceLastPrint)
				start = time.Now()
				bytesSinceLastPrint = 0
			}
		}
	}
	err = file.Sync()
	if err != nil {
		return err
	}

	return verify()
}

func printProgress(log *logrus.Entry, totalBytesWritten, partitionSize int64, start time.Time, bytesSinceLastPrint int64) {
	// Calculate progress and ETA
	progress := float64(totalBytesWritten) / float64(partitionSize) * 100
	elapsed := time.Since(start).Seconds()
	speed := float64(bytesSinceLastPrint) / elapsed                                   // Speed in bytes per second
	remainingSeconds := (float64(partitionSize) - float64(totalBytesWritten)) / speed // Remaining time in seconds
	remainingHours := float64(remainingSeconds / 3600)
	mbPerSecond := speed / (1024 * 1024)
	log.WithFields(logrus.Fields{
		"progress":  fmt.Sprintf("%.2f%%", progress),
		"speed":     fmt.Sprintf("%.2f MB/s", mbPerSecond),
		"remaining": fmt.Sprintf("%.2f hour(s)", remainingHours),
	}).Debug("")
}

// SetQuiet lowers the verbosity
func (z *FillZero) SetQuiet() {
	z.Quiet = true
}
