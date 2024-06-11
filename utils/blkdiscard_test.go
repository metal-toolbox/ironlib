package utils

import (
	"context"
	"os"
	"testing"

	"github.com/bmc-toolbox/common"
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
