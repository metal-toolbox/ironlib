package utils

import (
	"context"
	"testing"

	"github.com/bmc-toolbox/common"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func Test_blkdiscard(t *testing.T) {
	logger, hook := test.NewNullLogger()
	defer hook.Reset()

	drive := &common.Drive{Common: common.Common{LogicalName: "/dev/sdZZZ"}}
	err := NewFakeBlkdiscard().WipeDrive(context.TODO(), logger, drive)
	assert.NoError(t, err)
}
