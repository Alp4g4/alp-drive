package folder

import (
	"context"
	"fmt"
	"uuid"

	"alp-drive/internal/application/ports/clock"
	"alp-drive/internal/domain/folder"
)

type moveFolderRepository interface {
	GetFolder(ctx context.Context, id uuid.UUID) (folder.Folder, error)
	ExistsFolderByID(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateFolder(ctx context.Context, f folder.Folder) error
}

type MoveFolderUC struct {
	folderRepo moveFolderRepository
	clock      clock.Clock
}

func NewMoveFolderUC(folderRepo moveFolderRepository, clk clock.Clock) *MoveFolderUC {
	return &MoveFolderUC{
		folderRepo: folderRepo,
		clock:      clk,
	}
}

type MoveFolderInput struct {
	FolderID    uuid.UUID
	NewParentID *uuid.UUID
}

func (i MoveFolderInput) Validate() error {
	if i.FolderID == uuid.Nil() {
		return ErrFolderIDEmpty
	}

	return nil
}

func (uc *MoveFolderUC) Execute(ctx context.Context, input MoveFolderInput) (folder.Folder, error) {
	zero := folder.Folder{}

	if err := input.Validate(); err != nil {
		return zero, err
	}

	f, err := uc.folderRepo.GetFolder(ctx, input.FolderID)
	if err != nil {
		return zero, fmt.Errorf("loading folder: %w", err)
	}

	if input.NewParentID != nil {
		exists, err := uc.folderRepo.ExistsFolderByID(ctx, *input.NewParentID)
		if err != nil {
			return zero, fmt.Errorf("checking parent folder: %w", err)
		}

		if !exists {
			return zero, ErrParentFolderNotFound
		}
	}

	if err := f.Move(input.NewParentID); err != nil {
		return zero, err
	}

	f.UpdatedAt = uc.clock.Now()

	if err := uc.folderRepo.UpdateFolder(ctx, f); err != nil {
		return zero, fmt.Errorf("updating folder: %w", err)
	}

	return f, nil
}
