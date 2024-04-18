package utils

import (
	"crypto/rand"
	"io"
	"log"
	"math/big"
	"os"
	"slices"
)

const (
	bufferSize    = 512
	numWatermarks = 10
)

type watermark struct {
	position uint64
	data     []byte
}

func PrepWatermarks(disk string) (func() bool, error) {
	log.Printf("%s | Initiating watermarking process", disk)
	watermarks := make([]watermark, numWatermarks)

	// Write open
	file, err := os.OpenFile(disk, os.O_WRONLY, FileMode644)
	if err != nil {
		log.Println("Failed to open :"+disk, err)
		return nil, err
	}
	defer file.Close()

	// Get disk or partition size
	fileSize, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// Write watermarks on random locations
	err = writeWatermarks(file, fileSize, watermarks)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	checker := func() bool {
		log.Printf("%s | Checking if the watermark has been removed", disk)

		file, err := os.OpenFile(disk, os.O_RDONLY, FileMode644)
		if err != nil {
			log.Printf("%s | Failed to open disk for reading: %s", disk, err)
			return false
		}
		defer file.Close()

		for _, watermark := range watermarks {
			_, err = file.Seek(int64(watermark.position), io.SeekStart)
			if err != nil {
				log.Println("Error moving rw pointer:", err)
				return false
			}
			// Read the watermark written to the position
			currentValue := make([]byte, bufferSize)
			_, err = file.Read(currentValue)
			if err != nil {
				log.Println("Failed to read from file:", err)
				return false
			}
			// Check if the watermark is still in the disk
			if slices.Equal(currentValue, watermark.data) {
				log.Println("We have an existing watermark in the file:", err)
				return false
			}
		}
		log.Printf("%s | Watermark has been removed", disk)
		return true
	}
	return checker, nil
}

func writeWatermarks(file *os.File, fileSize int64, watermarks []watermark) error {
	for i := 0; i < numWatermarks; i++ {
		data := make([]byte, bufferSize)
		_, err := rand.Read(data)
		if err != nil {
			log.Println("error:", err)
			return err
		}
		randomPosition, err := rand.Int(rand.Reader, big.NewInt(fileSize))
		if err != nil {
			log.Println(err)
			return err
		}

		_, err = file.Seek(int64(randomPosition.Uint64()), io.SeekStart)
		if err != nil {
			log.Println("Error moving rw pointer:", err)
			return err
		}
		watermarks[i].position = randomPosition.Uint64()
		watermarks[i].data = data
	}
	return nil
}
