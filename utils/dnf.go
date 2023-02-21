package utils

import (
	"context"
	"os"

	"text/template"

	"github.com/metal-toolbox/ironlib/model"
)

var (
	DellRepoTemplate = `[{{ .Name }}-{{ .RepoVersion }}-system-update_independent]
name={{ .Name}}-{{ .RepoVersion }}
baseurl={{ .BaseURL }}/os_independent/
{{- if .GPGCheck }}
gpgcheck=1
{{- else }}
gpgcheck=0
{{- end }}
# GPG keys are placed here when the ironlib docker image is built
gpgkey=file:///usr/libexec/dell_dup/0x756ba70b1019ced6.asc
		file:///usr/libexec/dell_dup/0x1285491434D8786F.asc
		file:///usr/libexec/dell_dup/0xca77951d23b66a9d.asc
		file:///usr/libexec/dell_dup/0x3CA66B4946770C59.asc
enabled=1
exclude=dell-system-update*.i386

[{{ .Name }}-{{ .RepoVersion }}-dell-system-update_dependent]
name={{ .Name }}-{{ .RepoVersion }}-system-update_dependent
baseurl={{ .BaseURL }}/os_dependent/RHEL8_64
{{- if .GPGCheck }}
gpgcheck=1
{{- else }}
gpgcheck=0
{{- end }}
# GPG keys are placed here when the ironlib docker image is built
gpgkey=file:///usr/libexec/dell_dup/0x756ba70b1019ced6.asc
		file:///usr/libexec/dell_dup/0x1285491434D8786F.asc
		file:///usr/libexec/dell_dup/0xca77951d23b66a9d.asc
		file:///usr/libexec/dell_dup/0x3CA66B4946770C59.asc
enabled=1`
)

const (
	// EnvDnfUtility - to override the utility path
	EnvDnfUtility = "IRONLIB_UTIL_DNF"
)

type Dnf struct {
	Executor Executor
}

type DnfRepoParams struct {
	GPGCheck    bool
	Name        string
	BaseURL     string
	RepoVersion string
}

// Return a new dnf executor
func NewDnf(trace bool) *Dnf {
	utility := "microdnf"

	// lookup env var for util
	if eVar := os.Getenv(EnvDnfUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Dnf{
		Executor: e,
	}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (d *Dnf) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := d.Executor.CheckExecutable()

	return "dnf", d.Executor.CmdPath(), er
}

// Returns a fake dnf instance for tests
func NewFakeDnf() *Dnf {
	return &Dnf{
		Executor: NewFakeExecutor("microdnf"),
	}
}

// AddRepo sets up a dnf repo file with the given template and params
//
// path: the directory where the repo file is created, default: "/etc/yum.repos.d/"
func (d *Dnf) AddRepo(path string, params *DnfRepoParams, tmpl []byte) (err error) {
	if path == "" {
		path = "/etc/yum.repos.d/"
	}

	if params.BaseURL == "" {
		return ErrRepositoryBaseURL
	}

	f, err := os.Create(path + "/" + params.Name + ".repo")
	if err != nil {
		return err
	}

	t, err := template.New("repo").Parse(DellRepoTemplate)
	if err != nil {
		return err
	}

	return t.Execute(f, params)
}

// Install given packages
func (d *Dnf) Install(pkgNames []string) (err error) {
	args := []string{"install", "-y"}
	args = append(args, pkgNames...)

	d.Executor.SetArgs(args)

	_, err = d.Executor.ExecWithContext(context.Background())

	return err
}
