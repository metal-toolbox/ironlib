package utils

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"testing"

	common "github.com/metal-toolbox/bmc-common"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func Test_NewFillZeroCmd(t *testing.T) {
	require.NotNil(t, NewFillZeroCmd(false))
}

func Test_FillZeroWipeDrive(t *testing.T) {
	for _, size := range []int64{8192 - 1, 8192, 8192 + 1, 8192 * 2} {
		t.Run(fmt.Sprintf("%d", size), func(t *testing.T) {
			// Create a temporary file for testing
			tmpfile, err := os.CreateTemp("", "example")
			require.NoError(t, err)
			defer os.Remove(tmpfile.Name()) // clean up

			// Write some content to the temporary file
			_, err = io.CopyN(tmpfile, rand.Reader, size)
			require.NoError(t, err)
			require.NoError(t, tmpfile.Sync())
			require.NoError(t, tmpfile.Close())

			// Sanity check to make sure file size is written as expected and does not contain all zeros
			fileInfo, err := os.Stat(tmpfile.Name())
			require.NoError(t, err)
			require.Equal(t, size, fileInfo.Size())

			f, err := os.Open(tmpfile.Name())
			require.NoError(t, err)

			var buf bytes.Buffer
			n, err := io.Copy(&buf, f)
			require.NoError(t, err)
			require.Equal(t, size, n)
			require.NotEqual(t, make([]byte, size), buf.Bytes())
			require.NoError(t, f.Close())

			// Simulate a context
			ctx := context.Background()

			zw := &FillZero{}
			drive := &common.Drive{Common: common.Common{LogicalName: tmpfile.Name()}}
			logger, hook := test.NewNullLogger()
			defer hook.Reset()

			err = zw.WipeDrive(ctx, logger, drive)
			require.NoError(t, err)

			// Check if the file size remains the same after overwrite
			fileInfo, err = os.Stat(tmpfile.Name())
			require.NoError(t, err)
			require.Equal(t, size, fileInfo.Size())

			// Verify contents are all zero
			f, err = os.Open(tmpfile.Name())
			require.NoError(t, err)

			buf.Reset()
			n, err = io.Copy(&buf, f)
			require.NoError(t, err)
			require.Equal(t, size, n)
			require.Equal(t, make([]byte, size), buf.Bytes())
		})
	}
}
