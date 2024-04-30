package utils

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/metal-toolbox/ironlib/model"
)

type UEFIVariableCollector struct{}

func (UEFIVariableCollector) Attributes() (model.CollectorUtility, string, error) {
	return "uefi-variable-collector", "", nil
}

type UEFIVarEntry struct {
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Sha256sum string `json:"sha256sum"`
	Error     bool   `json:"error"`
}

type UEFIVars map[string]UEFIVarEntry

func (UEFIVariableCollector) GetUEFIVars(ctx context.Context) (UEFIVars, error) {
	uefivars := make(map[string]UEFIVarEntry)
	walkme := "/sys/firmware/efi/efivars"
	err := filepath.Walk(walkme, func(path string, info fs.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		entry := UEFIVarEntry{Path: path}
		if err != nil {
			// Capture all errors, even directories
			entry.Error = true
			uefivars[info.Name()] = entry
			return nil //nolint:nilerr // Not an error, keep walking
		}
		// No need to capture anything for directory entries without errors
		if info.IsDir() {
			return nil
		}
		entry.Size = info.Size()
		b, err := os.ReadFile(path)
		if err != nil {
			entry.Error = true
		} else {
			entry.Sha256sum = fmt.Sprintf("%x", sha256.Sum256(b))
		}
		uefivars[info.Name()] = entry
		return nil // Keep walking
	})
	if err != nil {
		return nil, err
	}
	return uefivars, nil
}
