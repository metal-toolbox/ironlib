package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/packethost/ironlib/model"
)

// downcases and returns a normalized vendor name from the given string
func FormatVendorName(v string) string {

	switch v {
	case "Dell Inc.":
		return "dell"
	case "HP", "HPE":
		return "hp"
	case "Supermicro":
		return "supermicro"
	case "Quanta Cloud Technology Inc.":
		return "quanta"
	case "GIGABYTE":
		return "gigabyte"
	case "Intel Corporation":
		return "intel"
	default:
		return v
	}
}

// Return a normalized product name given a product name
func FormatProductName(s string) string {
	switch s {
	case "SSG-6029P-E1CR12L-PH004":
		return "SSG-6029P-E1CR12L"
	case "SYS-5019C-MR-PH004":
		return "SYS-5019C-MR"
	default:
		return s
	}
}

// Return the product vendor name, given a product name/model string
func vendorFromString(s string) string {
	switch {
	case strings.Contains(s, "LSI3008-IT"):
		return "LSI"
	case strings.Contains(s, "HGST "):
		return "HGST"
	case strings.Contains(s, "Micron_"):
		return "Micron"
	case strings.Contains(s, "TOSHIBA"):
		return "Toshiba"
	case strings.Contains(s, "ConnectX4LX"):
		return "Mellanox"
	default:
		return "unknown"
	}
}

// return the given string with the idx prefixed
func prefixIndex(idx int, s string) string {
	return fmt.Sprintf("[%d] %s", idx, s)
}

func purgeTestComponentID(components []*model.Component) []*model.Component {
	for _, c := range components {
		c.ID = ""
	}
	return components
}

// Retrieve update file from updateFileURL, validate checksum
// on success - returns the local filesystem path to the update file that was retrieved and checksummed
func RetrieveUpdateFile(updateFileURL, targetDir string) (string, error) {

	if updateFileURL == "" {
		return "", fmt.Errorf("expected valid updateFileURL, got empty")
	}

	// the sha1 sum for each update file is expected to be present under the same UpdateFileURL with the .sha1 suffix
	checksumURL := updateFileURL + ".sha1"

	// updateFileURL to dest file name map
	m := map[string]string{
		updateFileURL: targetDir + "/" + path.Base(updateFileURL),
		checksumURL:   targetDir + "/" + path.Base(checksumURL),
	}

	// fetch update file
	for url, dstFile := range m {
		err := FetchFile(url, dstFile)
		if err != nil {
			return "", fmt.Errorf("file retrieve error, url: %s, err: %s", url, err)
		}
	}

	// validate checksum
	err := ValidateSHA1Checksum(m[updateFileURL], m[checksumURL])
	if err != nil {
		return "", fmt.Errorf("checksum error, file: %s, err: %s", m[updateFileURL], err.Error())
	}

	return m[updateFileURL], nil
}

// fetch file from the url and write to filePath
func FetchFile(fileURL, filePath string) error {

	// create file
	fileHandle, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer fileHandle.Close()

	// fetch url
	client := http.Client{Timeout: 180 * time.Second}
	resp, err := client.Get(fileURL)
	if err != nil {
		return err
	}

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response %s", resp.Status)
	}

	defer resp.Body.Close()

	// write response to file - io Copy reads in 32kb chunks so mem consumption shouldn't be a worry
	_, err = io.Copy(fileHandle, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// Returns error if SHA1 checksum is in valid
// filePath is the file of which checksum is to be validated
// sha1ChecksumFile is the file to read the checksum from
func ValidateSHA1Checksum(filePath, sha1ChecksumFile string) error {

	// read in the checksum from the checksum file
	expectedBytes, err := ioutil.ReadFile(sha1ChecksumFile)
	if err != nil {
		return err
	}

	expectedTokens := strings.Fields(string(expectedBytes))
	if len(expectedTokens) == 0 {
		return fmt.Errorf("checksum file invalid: %s", sha1ChecksumFile)
	}

	expected := expectedTokens[0]

	// open target file, truncating any existing
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer f.Close()

	// compute sha1 hash of target file
	hash := sha1.New()
	_, err = io.Copy(hash, f)
	if err != nil {
		return err
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	if checksum != expected {
		return fmt.Errorf("invalid checksum, expected %s, got %s", expected, checksum)
	}

	return nil
}

// compares the newVersion string with the oldVersion version and returns bool
func VersionIsNewer(newVersion, oldVersion string) (bool, error) {

	// validate semver versions
	newV, err := version.NewVersion(newVersion)
	if err != nil {
		return false, fmt.Errorf("semver version error: " + err.Error())
	}

	oldV, err := version.NewVersion(oldVersion)
	if err != nil {
		return false, fmt.Errorf("semver version error: " + err.Error())
	}

	return newV.GreaterThan(oldV), nil
}
