package utils

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"testing"

	common "github.com/metal-toolbox/bmc-common"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func Test_blkdiscard(t *testing.T) {
	logger, hook := test.NewNullLogger()
	defer hook.Reset()

	d := t.TempDir()
	tmpfile, err := os.CreateTemp(d, "fake-block-device")
	assert.NoError(t, err)

	err = os.WriteFile(tmpfile.Name(), make([]byte, 8192), 0o600)
	assert.NoError(t, err)

	drive := &common.Drive{Common: common.Common{LogicalName: tmpfile.Name()}}
	err = NewFakeBlkdiscard().WipeDrive(context.TODO(), logger, drive)

	// fake-block-device isn't a blockdevice that supports TRIM so we expect an error
	assert.Error(t, err)
}

// Test blkdiscard using a file mounted as a loopback device. This test requires
// root privileges, and is only run if the ENABLE_PRIV_TESTS environment variable
// is set to true.
func Test_blkdiscard_loopback(t *testing.T) {
	if os.Getenv("ENABLE_PRIV_TESTS") != "true" {
		t.Skip("Skipping privileged test blkdiscard_loopback")
	}

	tempDir := t.TempDir()

	// Create a 10 MB temporary file
	fileSize := int64(10 * 1024 * 1024)
	fileName := tempDir + "testfile"
	knownOdOutput := `0000000 000000 000000 000000 000000 000000 000000 000000 000000
*
50000000
`
	f, err := os.Create(fileName)
	assert.NoError(t, err)

	// Fill the file with 1's
	_, err = f.WriteAt(bytes.Repeat([]byte{1}, int(fileSize)), 0)
	f.Close()
	assert.NoError(t, err)
	t.Cleanup(func() { os.Remove(fileName) })

	// Create a loopback device using losetup
	losetupOutput, err := exec.Command("losetup", "--show", "-f", fileName).Output()
	loopbackDevice := string(bytes.TrimSpace(losetupOutput))
	t.Logf("loopbackDevice is: %s", loopbackDevice)
	assert.NoError(t, err)
	t.Cleanup(func() {
		cleanerr := exec.Command("losetup", "--detach", loopbackDevice).Run()
		if cleanerr != nil {
			println("WARNING: trouble cleaning up loopback device")
		}
	})

	// Make note of the sha256 checksum of the block device
	sha256Output, err := exec.Command("sha256sum", loopbackDevice).Output()
	t.Logf("sha256sum output before wipe is: %s", string(sha256Output))
	assert.NoError(t, err)
	sha256Before := string(bytes.Fields(sha256Output)[0])

	// Run blkdiscard on the loopback device
	d := NewBlkdiscardCmd(false)
	err = d.Discard(context.TODO(), &common.Drive{Common: common.Common{LogicalName: loopbackDevice}})
	assert.NoError(t, err)

	// Make note of the sha256 checksum of the block device after running blkdiscard
	sha256Output, err = exec.Command("sha256sum", loopbackDevice).Output()
	assert.NoError(t, err)
	t.Logf("sha256sum output after wipe is: %s", string(sha256Output))
	sha256After := string(bytes.Fields(sha256Output)[0])

	// The sha256 checksum should be different after running blkdiscard
	assert.NotEqual(t, sha256Before, sha256After)

	// od the file for funsies
	odOutput, err := exec.Command("od", loopbackDevice).Output()
	assert.NoError(t, err)
	t.Logf("od output after wipe is: %s", string(odOutput))
	assert.Equal(t, string(odOutput), knownOdOutput)
}
