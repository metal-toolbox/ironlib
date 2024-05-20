package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

// FakeExecute implements the utils.Executor interface
// to enable testing
type FakeExecute struct {
	Cmd      string
	Args     []string
	Env      []string
	CheckBin bool
	Stdin    io.Reader
	Stdout   []byte // Set this for the dummy data to be returned
	Stderr   []byte // Set this for the dummy data to be returned
	Quiet    bool
	ExitCode int
}

func NewFakeExecutor(cmd string) Executor {
	return &FakeExecute{Cmd: cmd, CheckBin: false}
}

// nolint:gocyclo // TODO: break this method up and move into each $util_test.go
// FakeExecute method returns whatever you want it to return
// Set e.Stdout and e.Stderr to data to be returned
func (e *FakeExecute) Exec(_ context.Context) (*Result, error) {
	switch e.Cmd {
	case "ipmicfg":
		if e.Args[0] == "-summary" {
			buf := new(bytes.Buffer)

			_, err := buf.ReadFrom(e.Stdin)
			if err != nil {
				return nil, err
			}

			e.Stdout = buf.Bytes()
		}
	case "mlxup":
	case "nvme":
		switch e.Args[0] {
		case "list":
			e.Stdout = nvmeListDummyJSON
		case "list-ns":
			e.Stdout = []byte(`{"nsid_list":[{"nsid":1}]}`)
		case "delete-ns":
			e.Stdout = []byte("delete-ns: Success, deleted nsid:1\n")
		case "create-ns":
			e.Stdout = []byte("create-ns: Success, created nsid:1\n")
		case "attach-ns":
			e.Stdout = []byte("attach-ns: Success, nsid:1\n")
		case "id-ns":
			e.Stdout = []byte(`{"lbafs":[{"ds":9},{"ds":12}]}`)
		case "reset", "ns-rescan":
		case "id-ctrl":
			cwd, _ := os.Getwd()
			f := "../fixtures/utils/nvme/nvmecli-id-ctrl"

			// This is a hack - until the fake executor can
			// live in its own along with the fixture, so the
			// fixture data does not have to be reached through a relative path.
			if strings.Contains(cwd, "providers") {
				f = "../../fixtures/utils/nvme/nvmecli-id-ctrl"
			}

			b, err := os.ReadFile(f)
			if err != nil {
				return nil, err
			}

			e.Stdout = b
		case "format", "sanitize":
			dev := e.Args[len(e.Args)-1]
			f, err := os.OpenFile(dev, os.O_WRONLY, 0)
			if err != nil {
				return nil, err
			}
			size, err := f.Seek(0, io.SeekEnd)
			if err != nil {
				return nil, err
			}
			err = f.Truncate(0)
			if err != nil {
				return nil, err
			}
			err = f.Sync()
			if err != nil {
				return nil, err
			}
			err = f.Truncate(size)
			if err != nil {
				return nil, err
			}
			err = f.Sync()
			if err != nil {
				return nil, err
			}
			err = f.Close()
			if err != nil {
				return nil, err
			}
		case "sanitize-log":
			dev := e.Args[len(e.Args)-1]
			dev = path.Base(dev)
			e.Stdout = []byte(fmt.Sprintf(`{%q:{"sprog":65535}}`, dev))
		}
	case "hdparm":
		if e.Args[0] == "-I" {
			cwd, _ := os.Getwd()

			f := "../fixtures/utils/hdparm/hdparm-i"

			// This is a hack - until the fake executor can
			// live in its own along with the fixture, so the
			// fixture data does not have to be reached through a relative path.
			if strings.Contains(cwd, "providers") {
				f = "../../fixtures/utils/hdparm/hdparm-i"
			}

			b, err := os.ReadFile(f)
			if err != nil {
				return nil, err
			}

			e.Stdout = b
		}
	case "lsblk":
		if e.Args[0] == "--json" {
			cwd, _ := os.Getwd()
			f := "../fixtures/utils/lsblk/json"

			// This is a hack - until the lsblk fake executor can
			// live in its own along with the fixture, so the
			// fixture data does not have to be reached through a relative path.
			if strings.Contains(cwd, "providers") {
				f = "../../fixtures/utils/lsblk/json"
			}

			b, err := os.ReadFile(f)
			if err != nil {
				return nil, err
			}

			e.Stdout = b
		}
	case "dsu":
	case "rpm":
		if e.Args[1] == "-1" && e.Args[2] == "dell-system-update" {
			e.Stdout = []byte("1.8.0-20.04.00")
		}
	case "msecli":
		if os.Getenv("FAIL_MICRON_UPDATE") != "" {
			return &Result{
				Stderr:   []byte("Folder /tmp/updates/Micron/D1MU020 is an invalid firmware update directory!"),
				ExitCode: 1,
			}, nil
		}

		if os.Getenv("FAIL_MICRON_QUERY") != "" {
			return &Result{
				Stdout:   []byte(``),
				ExitCode: 0,
			}, nil
		}
	}

	return &Result{Stdout: e.Stdout, Stderr: e.Stderr, ExitCode: 0}, nil
}

// CheckExecutable implements the Executor interface
func (e *FakeExecute) CheckExecutable() error {
	return nil
}

// CmdPath returns the absolute path to the executable
// this means the caller should not have disabled CheckBin.
func (e *FakeExecute) CmdPath() string {
	return e.Cmd
}

func (e *FakeExecute) SetArgs(a ...string) {
	e.Args = a
}

func (e *FakeExecute) SetEnv(env []string) {
	e.Env = env
}

func (e *FakeExecute) SetQuiet() {
	e.Quiet = true
}

func (e *FakeExecute) SetVerbose() {
	e.Quiet = false
}

func (e *FakeExecute) SetStdout(b []byte) {
	e.Stdout = b
}

func (e *FakeExecute) SetStderr(b []byte) {
	e.Stderr = b
}

func (e *FakeExecute) SetStdin(r io.Reader) {
	e.Stdin = r
}

func (e *FakeExecute) DisableBinCheck() {
	e.CheckBin = false
}

func (e *FakeExecute) SetExitCode(i int) {
	e.ExitCode = i
}

func (e *FakeExecute) GetCmd() string {
	cmd := []string{e.Cmd}
	cmd = append(cmd, e.Args...)

	return strings.Join(cmd, " ")
}

var nvmeListDummyJSON = []byte(`{
		"Devices" : [
		  {
			"NameSpace" : 1,
			"DevicePath" : "/dev/nvme0n1",
			"Firmware" : "AGGA4104",
			"Index" : 0,
			"ModelNumber" : "KXG60ZNV256G TOSHIBA",
			"ProductName" : "NULL",
			"SerialNumber" : "Z9DF70I8FY3L",
			"UsedBytes" : 256060514304,
			"MaximumLBA" : 500118192,
			"PhysicalSize" : 256060514304,
			"SectorSize" : 512
		  },
		  {
			"NameSpace" : 1,
			"DevicePath" : "/dev/nvme1n1",
			"Firmware" : "AGGA4104",
			"Index" : 1,
			"ModelNumber" : "KXG60ZNV256G TOSHIBA",
			"ProductName" : "NULL",
			"SerialNumber" : "Z9DF70I9FY3L",
			"UsedBytes" : 256060514304,
			"MaximumLBA" : 500118192,
			"PhysicalSize" : 256060514304,
			"SectorSize" : 512
		  }
		]
	  }
	`)
