package utils

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/bmc-toolbox/common"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func Test_NewFillZeroCmd(t *testing.T) {
	require.NotNil(t, NewFillZeroCmd(false))
}

func Test_WipeDrive(t *testing.T) {
	for _, size := range []int64{4095, 4096, 4097, 8192} {
		t.Run(fmt.Sprintf("%d", size), func(t *testing.T) {
			// Create a temporary file for testing
			tmpfile, err := os.CreateTemp("", "example")
			require.NoError(t, err)
			defer os.Remove(tmpfile.Name()) // clean up

			// Write some content to the temporary file
			_, err = tmpfile.Write(make([]byte, size))
			require.NoError(t, err)

			// Simulate a context
			ctx := context.Background()

			// Create a FillZero instance
			zw := &FillZero{}
			drive := &common.Drive{Common: common.Common{LogicalName: tmpfile.Name()}}

			// Test Fill function
			logger, hook := test.NewNullLogger()
			defer hook.Reset()
			err = zw.WipeDrive(ctx, logger, drive)
			require.NoError(t, err)

			// Check if the file size remains the same after overwrite
			fileInfo, err := os.Stat(tmpfile.Name())
			require.NoError(t, err)
			require.Equal(t, size, fileInfo.Size())
		})
	}
}
