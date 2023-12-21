package utils

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
)

type EfiVars struct {
	Vars []*Entry `json:"vars"`
}
type Entry struct {
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Sha256sum string `json:"sha256sum"`
	Error     bool   `json:"error"`
}

func Example() {
	var efivars EfiVars
	walkme := "/sys/firmware/efi/efivars"
	_ = filepath.Walk(walkme, func(path string, info fs.FileInfo, err error) error {
		entry := Entry{Path: path}
		if err != nil {
			// Capture all errors, even directories
			entry.Error = true
			efivars.Vars = append(efivars.Vars, &entry)
			return nil // Keep walking
		}
		// No need to capture anything for directory entries without errors
		if info.IsDir() {
			return nil
		}
		entry.Size = info.Size()
		b, err := ioutil.ReadFile(path)
		if err != nil {
			entry.Error = true
		} else {
			entry.Sha256sum = fmt.Sprintf("%x", sha256.Sum256(b))
		}
		efivars.Vars = append(efivars.Vars, &entry)
		return nil // Keep walking
	})
	data, err := json.Marshal(efivars)
	if err != nil {
		fmt.Println("sad2 ", err)
	} else {
		fmt.Println(string(data))
	}
}
