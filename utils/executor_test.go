package utils

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"testing"

	"github.com/packethost/ironlib/errs"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func Test_Stdin(t *testing.T) {
	e := new(Execute)
	e.Cmd = "grep"
	e.Args = []string{"hello"}
	e.Stdin = bytes.NewReader([]byte("hello"))
	e.SetQuiet()

	result, err := e.ExecWithContext(context.Background())
	if err != nil {
		fmt.Println(err.Error())
	}

	assert.Equal(t, []byte("hello\n"), result.Stdout)
}

type checkBinTester struct {
	createFile  bool
	filePath    string
	expectedErr error
	fileMode    uint
	testName    string
}

func initCheckBinTests() []checkBinTester {
	return []checkBinTester{
		{
			false,
			"f",
			errs.ErrBinLookupPath,
			0,
			"bin path lookup err test",
		},
		{
			false,
			"/tmp/f",
			errs.ErrBinLstat,
			0,
			"bin exists err test",
		},
		{
			true,
			"/tmp/f",
			errs.ErrBinNotExecutable,
			0666,
			"bin exists with no executable bit test",
		},
		{
			true,
			"/tmp/j",
			nil,
			0667,
			"bin with executable bit returns no error",
		},
		{
			true,
			"/tmp/k",
			nil,
			0700,
			"bin with owner executable bit returns no error",
		},
		{
			true,
			"/tmp/l",
			nil,
			0070,
			"bin with group executable bit returns no error",
		},
		{
			true,
			"/tmp/m",
			nil,
			0007,
			"bin with other executable bit returns no error",
		},
	}
}

func Test_checkBinDep(t *testing.T) {
	tests := initCheckBinTests()
	for _, c := range tests {
		if c.createFile {
			f, err := os.Create(c.filePath)
			if err != nil {
				t.Error(err)
			}

			defer os.Remove(c.filePath)

			if c.fileMode != 0 {
				err = f.Chmod(fs.FileMode(c.fileMode))
				if err != nil {
					t.Error(err)
				}
			}
		}

		err := checkBinDep(c.filePath)
		assert.Equal(t, c.expectedErr, errors.Cause(err), c.testName)
	}
}
