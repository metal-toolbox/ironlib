package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AddRepo(t *testing.T) {
	params := &DnfRepoParams{
		GPGCheck:    true,
		Name:        "dell",
		BaseURL:     "http://example.fup/firmware/fup/dell/1.0.1",
		RepoVersion: "1.0.1",
	}

	expected := []byte(`[dell-1.0.1-system-update_independent]
name=dell-1.0.1
baseurl=http://example.fup/firmware/fup/dell/1.0.1/os_independent/
gpgcheck=1
# GPG keys are placed here when the ironlib docker image is built
gpgkey=file:///usr/libexec/dell_dup/0x756ba70b1019ced6.asc
		file:///usr/libexec/dell_dup/0x1285491434D8786F.asc
		file:///usr/libexec/dell_dup/0xca77951d23b66a9d.asc
		file:///usr/libexec/dell_dup/0x3CA66B4946770C59.asc
enabled=1
exclude=dell-system-update*.i386

[dell-1.0.1-dell-system-update_dependent]
name=dell-1.0.1-system-update_dependent
baseurl=http://example.fup/firmware/fup/dell/1.0.1/os_dependent/RHEL8_64
gpgcheck=1
# GPG keys are placed here when the ironlib docker image is built
gpgkey=file:///usr/libexec/dell_dup/0x756ba70b1019ced6.asc
		file:///usr/libexec/dell_dup/0x1285491434D8786F.asc
		file:///usr/libexec/dell_dup/0xca77951d23b66a9d.asc
		file:///usr/libexec/dell_dup/0x3CA66B4946770C59.asc
enabled=1`)

	dnf := NewFakeDnf()

	err := dnf.AddRepo("/tmp/", params, []byte(DellRepoTemplate))
	if err != nil {
		t.Error(err)
	}

	b, err := os.ReadFile("/tmp/" + params.Name + ".repo")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, string(expected), string(b))
}
