package folder

import (
	"context"
	"fmt"
	"strings"
	"uuid"

	"alp-drive/internal/application/ports/clock"
	"alp-drive/internal/domain/folder"
)

type createFolderRepository interface {
	ExistsFolderByID(ctx context.Context, id uuid.UUID) (bool, error)
	CreateFolder(ctx context.Context, f folder.Folder) error
}

type CreateFolderUC struct {
	folderRepo createFolderRepository
	clock      clock.Clock
}

func NewCreateFolderUC(folderRepo createFolderRepository, clk clock.Clock) *CreateFolderUC {
	return &CreateFolderUC{
		folderRepo: folderRepo,
		clock:      clk,
	}
}

type CreateFolderInput struct {
	Name     string
	ParentID *uuid.UUID
}

func (i CreateFolderInput) Validate() error {
	if strings.TrimSpace(i.Name) == "" {
		return ErrFolderNameEmpty
	}

	return nil
}

func (uc *CreateFolderUC) Execute(ctx context.Context, input CreateFolderInput) (folder.Folder, error) {
	zero := folder.Folder{}

	if err := input.Validate(); err != nil {
		return zero, err
	}

	if input.ParentID != nil {
		exists, err := uc.folderRepo.ExistsFolderByID(ctx, *input.ParentID)
		if err != nil {
			return zero, fmt.Errorf("checking parent folder: %w", err)
		}

		if !exists {
			return zero, ErrParentFolderNotFound
		}
	}

	now := uc.clock.Now()

	f := folder.Folder{
		ID:        uuid.NewV7(),
		Name:      input.Name,
		Parent:    input.ParentID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.folderRepo.CreateFolder(ctx, f); err != nil {
		return zero, fmt.Errorf("creating folder: %w", err)
	}

	return f, nil
}
