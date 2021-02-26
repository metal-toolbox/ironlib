package utils

import (
	"bytes"
	"context"
	"fmt"
	"testing"

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
