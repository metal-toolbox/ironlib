package utils

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// FakeExecute implements the utils.Executor interface
// to enable testing
type FakeExecute struct {
	Cmd    string
	Args   []string
	Env    []string
	Stdin  io.Reader
	Stdout []byte // Set this for the dummy data to be returned
	Stderr []byte // Set this for the dummy data to be returned
	Quiet  bool
}

func NewFakeExecutor(cmd string) Executor {
	return &FakeExecute{Cmd: cmd}
}

// nolint:gocyclo // TODO: break this method up and move into each $util_test.go
// FakeExecute method returns whatever you want it to return
// Set e.Stdout and e.Stderr to data to be returned
func (e *FakeExecute) ExecWithContext(ctx context.Context) (*Result, error) {
	switch e.Cmd {
	case "ipmicfg":
		if e.Args[0] == "-summary" {
			b, err := ioutil.ReadFile("test_data/ipmicfg_summary")
			if err != nil {
				return nil, err
			}

			e.Stdout = b
		}
	case "mlxup":
		if e.Args[0] == "--query" {
			b, err := ioutil.ReadFile("test_data/mlxup_query")
			if err != nil {
				return nil, err
			}

			e.Stdout = b
		}
	case "nvme":
		if e.Args[0] == "list" {
			e.Stdout = nvmeListDummyJSON
			break
		}
	case "storecli":
		b, err := ioutil.ReadFile("test_data/storecli.json")
		if err != nil {
			return nil, err
		}

		e.Stdout = b
	case "storecli-nocontrollers":
		b, err := ioutil.ReadFile("test_data/storecli_nocontrollers.json")
		if err != nil {
			return nil, err
		}

		e.Stdout = b
	case "smartctl":
		switch e.Args[0] {
		case "--scan":
			b, err := ioutil.ReadFile("test_data/smartctl_scan.json")
			if err != nil {
				return nil, err
			}

			e.Stdout = b
		case "-a":
			if strings.Join(e.Args, " ") == "-a /dev/sda -j" {
				b, err := ioutil.ReadFile("test_data/smartctl_a_sda1.json")
				if err != nil {
					return nil, err
				}

				e.Stdout = b
			}

			if strings.Join(e.Args, " ") == "-a /dev/nvme0 -j" {
				b, err := ioutil.ReadFile("test_data/smartctl_a_nvme0.json")
				if err != nil {
					return nil, err
				}

				e.Stdout = b
			}
		}
	case "dsu":
		if e.Args[1] == "--inventory" {
			b, err := ioutil.ReadFile("test_data/r6515/dsu_inventory")
			if err != nil {
				return nil, err
			}

			e.Stdout = b
		}

		if e.Args[1] == "--preview" {
			b, err := ioutil.ReadFile("test_data/dsu_preview")
			if err != nil {
				return nil, err
			}

			e.Stdout = b
		}
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

		b, err := ioutil.ReadFile("test_data/msecli_list")
		if err != nil {
			return nil, err
		}

		e.Stdout = b
	}

	return &Result{Stdout: e.Stdout, Stderr: e.Stderr, ExitCode: 0}, nil
}

func (e *FakeExecute) SetArgs(a []string) {
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

func (e *FakeExecute) GetCmd() string {
	cmd := []string{e.Cmd}
	cmd = append(cmd, e.Args...)

	return strings.Join(cmd, " ")
}

var (
	nvmeListDummyJSON = []byte(`{
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
)
