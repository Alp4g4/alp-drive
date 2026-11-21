package file

import (
	"context"
	"fmt"
	"uuid"

	"alp-drive/internal/domain/file"
)

type deleteFileRepository interface {
	GetFile(ctx context.Context, id uuid.UUID) (file.File, error)
	DeleteFile(ctx context.Context, id uuid.UUID) error
}

type fileDeleter interface {
	Delete(ctx context.Context, fileID uuid.UUID) error
}

type DeleteFileUC struct {
	fileRepo    deleteFileRepository
	fileDeleter fileDeleter
}

func NewDeleteFileUC(fileRepo deleteFileRepository, fileDeleter fileDeleter) *DeleteFileUC {
	return &DeleteFileUC{
		fileRepo:    fileRepo,
		fileDeleter: fileDeleter,
	}
}

type DeleteFileInput struct {
	FileID uuid.UUID
}

func (i DeleteFileInput) Validate() error {
	if i.FileID == uuid.Nil() {
		return ErrFileIDEmpty
	}

	return nil
}

func (uc *DeleteFileUC) Execute(ctx context.Context, input DeleteFileInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	f, err := uc.fileRepo.GetFile(ctx, input.FileID)
	if err != nil {
		return fmt.Errorf("loading file: %w", err)
	}

	if err := uc.fileRepo.DeleteFile(ctx, f.ID); err != nil {
		return fmt.Errorf("deleting file: %w", err)
	}

	_ = uc.fileDeleter.Delete(ctx, f.ID)

	return nil
}
