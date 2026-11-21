package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"uuid"

	usecasefile "alp-drive/internal/application/usecases/file"
	"alp-drive/internal/domain/file"
	"alp-drive/internal/infrastructure/sqlite/sqlcgen"
)

func (s *Store) GetFile(ctx context.Context, id uuid.UUID) (file.File, error) {
	m, err := s.queries.GetFile(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return file.File{}, usecasefile.ErrFileNotFound
	}
	if err != nil {
		return file.File{}, fmt.Errorf("selecting file %s: %w", id, err)
	}

	return file.File(m), nil
}

func (s *Store) CreateFile(ctx context.Context, f file.File) error {
	if err := s.queries.CreateFile(ctx, sqlcgen.CreateFileParams(f)); err != nil {
		return fmt.Errorf("inserting file %s: %w", f.ID, err)
	}

	return nil
}

func (s *Store) UpdateFile(ctx context.Context, f file.File) error {
	rows, err := s.queries.UpdateFile(ctx, sqlcgen.UpdateFileParams{
		Name:      f.Name,
		Folder:    f.Folder,
		Size:      f.Size,
		MimeType:  f.MimeType,
		UpdatedAt: f.UpdatedAt,
		ID:        f.ID,
	})
	if err != nil {
		return fmt.Errorf("updating file %s: %w", f.ID, err)
	}

	if rows == 0 {
		return usecasefile.ErrFileNotFound
	}

	return nil
}

func (s *Store) DeleteFile(ctx context.Context, id uuid.UUID) error {
	rows, err := s.queries.DeleteFile(ctx, id)
	if err != nil {
		return fmt.Errorf("deleting file %s: %w", id, err)
	}

	if rows == 0 {
		return usecasefile.ErrFileNotFound
	}

	return nil
}

func (s *Store) ListFilesByFolderID(ctx context.Context, folderID uuid.UUID) ([]file.File, error) {
	rows, err := s.queries.ListFilesByFolderID(ctx, sqlcgen.ListFilesByFolderIDParams{
		Folder:     folderID,
		LimitCount: listLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("listing files of folder %s: %w", folderID, err)
	}

	files := make([]file.File, len(rows))
	for i, m := range rows {
		files[i] = file.File(m)
	}

	return files, nil
}
