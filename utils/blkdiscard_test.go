package utils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_blkdiscard(t *testing.T) {
	err := NewFakeBlkdiscard().Discard(context.TODO(), "/dev/sdZZZ")
	assert.NoError(t, err)
}
