package folder

import (
	"context"
	"fmt"
	"uuid"

	"alp-drive/internal/domain/folder"
)

type listFoldersRepository interface {
	ListFoldersByParentID(ctx context.Context, parentID *uuid.UUID) ([]folder.Folder, error)
}

type ListFoldersUC struct {
	folderRepo listFoldersRepository
}

func NewListFoldersUC(folderRepo listFoldersRepository) *ListFoldersUC {
	return &ListFoldersUC{
		folderRepo: folderRepo,
	}
}

type ListFoldersInput struct {
	ParentID *uuid.UUID
}

func (uc *ListFoldersUC) Execute(ctx context.Context, input ListFoldersInput) ([]folder.Folder, error) {
	folders, err := uc.folderRepo.ListFoldersByParentID(ctx, input.ParentID)
	if err != nil {
		return nil, fmt.Errorf("listing folders: %w", err)
	}

	return folders, nil
}
