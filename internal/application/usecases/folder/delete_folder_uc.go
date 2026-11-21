package folder

import (
	"context"
	"fmt"
	"uuid"

	"alp-drive/internal/domain/file"
)

type deleteFolderRepository interface {
	ExistsFolderByID(ctx context.Context, id uuid.UUID) (bool, error)
	DeleteFolderCascade(ctx context.Context, id uuid.UUID) ([]file.File, error)
}

type fileDeleter interface {
	Delete(ctx context.Context, fileID uuid.UUID) error
}

type DeleteFolderUC struct {
	folderRepo  deleteFolderRepository
	fileDeleter fileDeleter
}

func NewDeleteFolderUC(folderRepo deleteFolderRepository, fileDeleter fileDeleter) *DeleteFolderUC {
	return &DeleteFolderUC{
		folderRepo:  folderRepo,
		fileDeleter: fileDeleter,
	}
}

type DeleteFolderInput struct {
	FolderID uuid.UUID
}

func (i DeleteFolderInput) Validate() error {
	if i.FolderID == uuid.Nil() {
		return ErrFolderIDEmpty
	}

	return nil
}

func (uc *DeleteFolderUC) Execute(ctx context.Context, input DeleteFolderInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	exists, err := uc.folderRepo.ExistsFolderByID(ctx, input.FolderID)
	if err != nil {
		return fmt.Errorf("checking folder: %w", err)
	}

	if !exists {
		return ErrFolderNotFound
	}

	deletedFiles, err := uc.folderRepo.DeleteFolderCascade(ctx, input.FolderID)
	if err != nil {
		return fmt.Errorf("deleting folder: %w", err)
	}

	for _, f := range deletedFiles {
		_ = uc.fileDeleter.Delete(ctx, f.ID)
	}

	return nil
}
