package storages

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage implements Storage for local disk.
type LocalStorage struct {
	root string
}

func NewLocalStorage(root string) *LocalStorage {
	return &LocalStorage{root: root}
}

func (s *LocalStorage) path(p string) string {
	return filepath.Join(s.root, p)
}

func (s *LocalStorage) Exists(_ context.Context, path string) (bool, error) {
	_, err := os.Stat(s.path(path))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *LocalStorage) Get(_ context.Context, path string) (io.ReadCloser, error) {
	return os.Open(s.path(path))
}

func (s *LocalStorage) Put(_ context.Context, path string, r io.Reader) error {
	full := s.path(path)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return err
	}
	f, err := os.Create(full)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (s *LocalStorage) Delete(_ context.Context, path string) error {
	return os.Remove(s.path(path))
}

func (s *LocalStorage) Size(_ context.Context, path string) (int64, error) {
	info, err := os.Stat(s.path(path))
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (s *LocalStorage) Copy(_ context.Context, src string, dst string) error {
	srcFile, err := os.Open(s.path(src))
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFull := s.path(dst)
	if err = os.MkdirAll(filepath.Dir(dstFull), 0755); err != nil {
		return err
	}
	dstFile, err := os.Create(dstFull)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (s *LocalStorage) Move(_ context.Context, src string, dst string) error {
	dstFull := s.path(dst)
	if err := os.MkdirAll(filepath.Dir(dstFull), 0755); err != nil {
		return err
	}
	return os.Rename(s.path(src), dstFull)
}

