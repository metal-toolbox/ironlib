package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Executor interface lets us implement dummy executors for tests
type Executor interface {
	ExecWithContext(context.Context) (*Result, error)
	SetArgs([]string)
	SetEnv([]string)
	SetQuiet()
	SetVerbose()
	GetCmd() string
	SetStdin(io.Reader)
	// for tests
	SetStdout([]byte)
	SetStderr([]byte)
}

func NewExecutor(cmd string) Executor {
	return &Execute{Cmd: cmd}
}

// An execute instace
type Execute struct {
	Cmd   string
	Args  []string
	Env   []string
	Stdin io.Reader
	Quiet bool
}

// The result of a command execution
type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

func (e *Execute) GetCmd() string {
	cmd := []string{e.Cmd}
	cmd = append(cmd, e.Args...)
	return strings.Join(cmd, " ")
}

func (e *Execute) SetArgs(a []string) {
	e.Args = a
}

func (e *Execute) SetEnv(env []string) {
	e.Env = env
}

func (e *Execute) SetQuiet() {
	e.Quiet = true
}

func (e *Execute) SetVerbose() {
	e.Quiet = false
}

func (e *Execute) SetStdin(r io.Reader) {
	e.Stdin = r
}

func (e *Execute) SetStdout(b []byte) {
}

func (e *Execute) SetStderr(b []byte) {
}

// wrapResult wraps command execution results and returns
func wrapResult(stdout, stderr []byte, exitCode int) *Result {
	return &Result{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
	}
}

func (e *Execute) ExecWithContext(ctx context.Context) (result *Result, err error) {

	cmd := exec.CommandContext(ctx, e.Cmd, e.Args...)
	cmd.Env = append(cmd.Env, e.Env...)
	cmd.Stdin = e.Stdin

	var stdoutBuf, stderrBuf bytes.Buffer
	if !e.Quiet {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	} else {
		cmd.Stderr = &stderrBuf
		cmd.Stdout = &stdoutBuf
	}

	if err := cmd.Run(); err != nil {
		return wrapResult(stdoutBuf.Bytes(), stderrBuf.Bytes(), cmd.ProcessState.ExitCode()),
			fmt.Errorf("error executing command: `%s %s`, err: %s", e.Cmd, strings.Join(e.Args, " "), err.Error())
	}

	return wrapResult(stdoutBuf.Bytes(), stderrBuf.Bytes(), cmd.ProcessState.ExitCode()), nil
}
