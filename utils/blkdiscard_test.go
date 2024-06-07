package utils

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func Test_blkdiscard(t *testing.T) {
	logger, hook := test.NewNullLogger()
	defer hook.Reset()

	err := NewFakeBlkdiscard().WipeDisk(context.TODO(), logger, "/dev/sdZZZ")
	assert.NoError(t, err)
}
