package file

import (
	"context"
	"fmt"
	"uuid"

	"alp-drive/internal/domain/file"
)

type listFilesRepository interface {
	ListFilesByFolderID(ctx context.Context, folderID uuid.UUID) ([]file.File, error)
}

type ListFilesByFolderUC struct {
	fileRepo listFilesRepository
}

func NewListFilesByFolderUC(fileRepo listFilesRepository) *ListFilesByFolderUC {
	return &ListFilesByFolderUC{
		fileRepo: fileRepo,
	}
}

type ListFilesByFolderInput struct {
	FolderID uuid.UUID
}

func (i ListFilesByFolderInput) Validate() error {
	if i.FolderID == uuid.Nil() {
		return ErrFolderIDEmpty
	}

	return nil
}

func (uc *ListFilesByFolderUC) Execute(ctx context.Context, input ListFilesByFolderInput) ([]file.File, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	files, err := uc.fileRepo.ListFilesByFolderID(ctx, input.FolderID)
	if err != nil {
		return nil, fmt.Errorf("listing files: %w", err)
	}

	return files, nil
}
