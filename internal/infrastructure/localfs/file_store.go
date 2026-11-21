package localfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"uuid"
)

type FileStore struct {
	dir string
}

func NewFileStore(dir string) *FileStore {
	return &FileStore{dir: dir}
}

func (s *FileStore) Store(ctx context.Context, fileID uuid.UUID, content io.Reader) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return fmt.Errorf("creating file storage directory: %w", err)
	}

	tmp := s.path(fileID) + ".tmp"
	dst, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", fileID, err)
	}

	if _, err := io.Copy(dst, content); err != nil {
		dst.Close()
		os.Remove(tmp)
		return fmt.Errorf("writing file %s: %w", fileID, err)
	}

	if err := dst.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("closing file %s: %w", fileID, err)
	}

	if err := os.Rename(tmp, s.path(fileID)); err != nil {
		return fmt.Errorf("finalizing file %s: %w", fileID, err)
	}

	return nil
}

func (s *FileStore) Open(ctx context.Context, fileID uuid.UUID) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	f, err := os.Open(s.path(fileID))
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", fileID, err)
	}

	return f, nil
}

func (s *FileStore) Delete(ctx context.Context, fileID uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	err := os.Remove(s.path(fileID))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("deleting file %s: %w", fileID, err)
	}

	return nil
}

func (s *FileStore) path(fileID uuid.UUID) string {
	return filepath.Join(s.dir, fileID.String())
}
