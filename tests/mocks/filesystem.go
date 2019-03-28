package mocks

import (
	"os"
	"path/filepath"

	"github.com/rkcloudchain/rksync/config"
)

// NewFSMock creates a new FileSystemMock instance
func NewFSMock(basedir string) *FileSystemMock {
	return &FileSystemMock{basedir}
}

// FileSystemMock mocks a file system
type FileSystemMock struct {
	baseDir string
}

// Create ...
func (m *FileSystemMock) Create(chainID, filename string) (config.File, error) {
	p := filepath.Join(m.baseDir, filename)
	dir := filepath.Dir(p)
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	f, err := os.Create(p)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// OpenFile ...
func (m *FileSystemMock) OpenFile(chainID, filename string, flag int, perm os.FileMode) (config.File, error) {
	p := filepath.Join(m.baseDir, filename)
	return os.OpenFile(p, flag, perm)
}

// Stat ...
func (m *FileSystemMock) Stat(chainID, filename string) (os.FileInfo, error) {
	p := filepath.Join(m.baseDir, filename)
	return os.Stat(p)
}