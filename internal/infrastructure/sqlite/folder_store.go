package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"uuid"

	usecasefolder "alp-drive/internal/application/usecases/folder"
	"alp-drive/internal/domain/file"
	"alp-drive/internal/domain/folder"
	"alp-drive/internal/infrastructure/sqlite/sqlcgen"
)

func (s *Store) ExistsFolderByID(ctx context.Context, id uuid.UUID) (bool, error) {
	n, err := s.queries.ExistsFolderByID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("checking folder %s: %w", id, err)
	}

	return n == 1, nil
}

func (s *Store) CreateFolder(ctx context.Context, f folder.Folder) error {
	if err := s.queries.CreateFolder(ctx, sqlcgen.CreateFolderParams(f)); err != nil {
		return fmt.Errorf("inserting folder %s: %w", f.ID, err)
	}

	return nil
}

func (s *Store) GetFolder(ctx context.Context, id uuid.UUID) (folder.Folder, error) {
	m, err := s.queries.GetFolder(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return folder.Folder{}, usecasefolder.ErrFolderNotFound
	}
	if err != nil {
		return folder.Folder{}, fmt.Errorf("selecting folder %s: %w", id, err)
	}

	return folder.Folder(m), nil
}

func (s *Store) UpdateFolder(ctx context.Context, f folder.Folder) error {
	rows, err := s.queries.UpdateFolder(ctx, sqlcgen.UpdateFolderParams{
		Name:      f.Name,
		Parent:    f.Parent,
		UpdatedAt: f.UpdatedAt,
		ID:        f.ID,
	})
	if err != nil {
		return fmt.Errorf("updating folder %s: %w", f.ID, err)
	}

	if rows == 0 {
		return usecasefolder.ErrFolderNotFound
	}

	return nil
}

func (s *Store) ListFoldersByParentID(ctx context.Context, parentID *uuid.UUID) ([]folder.Folder, error) {
	rows, err := s.queries.ListFoldersByParentID(ctx, sqlcgen.ListFoldersByParentIDParams{
		Parent:     parentID,
		LimitCount: listLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("listing folders: %w", err)
	}

	folders := make([]folder.Folder, len(rows))
	for i, m := range rows {
		folders[i] = folder.Folder(m)
	}

	return folders, nil
}

func (s *Store) DeleteFolderCascade(ctx context.Context, id uuid.UUID) ([]file.File, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)

	rows, err := q.ListFilesByFolderTree(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("listing files under folder %s: %w", id, err)
	}

	deletedRows, err := q.DeleteFolder(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("deleting folder %s: %w", id, err)
	}
	if deletedRows == 0 {
		return nil, usecasefolder.ErrFolderNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing folder deletion %s: %w", id, err)
	}

	files := make([]file.File, len(rows))
	for i, m := range rows {
		files[i] = file.File(m)
	}

	return files, nil
}
