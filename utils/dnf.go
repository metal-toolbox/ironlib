package utils

import (
	"context"
	"fmt"
	"strings"
)

type Dnf struct {
	Executor Executor
}

// Return a new dnf executor
func NewDnf(trace bool) *Dnf {

	e := NewExecutor("dnf")
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})
	if !trace {
		e.SetQuiet()
	}

	return &Dnf{
		Executor: e,
	}
}

// Returns a fake dnf instance for tests
func NewFakeDnf() *Dnf {
	return &Dnf{
		Executor: NewFakeExecutor("dnf"),
	}
}

// Enable the given slice of repo names
func (d *Dnf) EnableRepo(repos []string) (err error) {

	for _, r := range repos {
		d.Executor.SetArgs([]string{"config-manager", "--enable", r})
		_, err = d.Executor.ExecWithContext(context.Background())
		if err != nil {
			return err
		}
	}

	return nil
}

// Attempt to install one of the given packages
func (d *Dnf) InstallOneOf(pNameVersions []string) (err error) {

	var installed bool
	for _, pkg := range pNameVersions {
		err := d.Install([]string{pkg})
		if err != nil {
			continue
		} else {
			installed = true
			break
		}
	}

	if !installed {
		return fmt.Errorf("failed to install one of: %s", strings.Join(pNameVersions, ","))
	}

	return nil
}

// Install given packages
func (d *Dnf) Install(pkgNames []string) (err error) {

	args := []string{"install", "-y"}
	for _, p := range pkgNames {
		args = append(args, p)
	}

	d.Executor.SetArgs(args)

	_, err = d.Executor.ExecWithContext(context.Background())
	if err != nil {
		return err
	}

	return nil
}
