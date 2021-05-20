package utils

import (
	"bytes"
	"context"
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

// GetCmd returns the command with args as a string
func (e *Execute) GetCmd() string {
	cmd := []string{e.Cmd}
	cmd = append(cmd, e.Args...)

	return strings.Join(cmd, " ")
}

// SetArgs sets the command args
func (e *Execute) SetArgs(a []string) {
	e.Args = a
}

// SetEnv sets the env variabls
func (e *Execute) SetEnv(env []string) {
	e.Env = env
}

// SetQuiet lowers the verbosity
func (e *Execute) SetQuiet() {
	e.Quiet = true
}

// SetVerbose does whats it says
func (e *Execute) SetVerbose() {
	e.Quiet = false
}

// SetStdin sets the reader to the command stdin
func (e *Execute) SetStdin(r io.Reader) {
	e.Stdin = r
}

// SetStdout doesn't do much, is around for tests
func (e *Execute) SetStdout(b []byte) {
}

// SetStderr doesn't do much, is around for tests
func (e *Execute) SetStderr(b []byte) {
}

// ExecWithContext executes the command and returns the Result object
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
		result = &Result{stdoutBuf.Bytes(), stderrBuf.Bytes(), cmd.ProcessState.ExitCode()}
		return result, newUtilsExecError(e.GetCmd(), result)
	}

	result = &Result{stdoutBuf.Bytes(), stderrBuf.Bytes(), cmd.ProcessState.ExitCode()}

	return result, nil
}
