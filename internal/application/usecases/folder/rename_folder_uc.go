package folder

import (
	"context"
	"fmt"
	"strings"
	"uuid"

	"alp-drive/internal/application/ports/clock"
	"alp-drive/internal/domain/folder"
)

type renameFolderRepository interface {
	GetFolder(ctx context.Context, id uuid.UUID) (folder.Folder, error)
	UpdateFolder(ctx context.Context, f folder.Folder) error
}

type RenameFolderUC struct {
	folderRepo renameFolderRepository
	clock      clock.Clock
}

func NewRenameFolderUC(folderRepo renameFolderRepository, clk clock.Clock) *RenameFolderUC {
	return &RenameFolderUC{
		folderRepo: folderRepo,
		clock:      clk,
	}
}

type RenameFolderInput struct {
	FolderID uuid.UUID
	Name     string
}

func (i RenameFolderInput) Validate() error {
	if i.FolderID == uuid.Nil() {
		return ErrFolderIDEmpty
	}

	if strings.TrimSpace(i.Name) == "" {
		return ErrFolderNameEmpty
	}

	return nil
}

func (uc *RenameFolderUC) Execute(ctx context.Context, input RenameFolderInput) (folder.Folder, error) {
	zero := folder.Folder{}

	if err := input.Validate(); err != nil {
		return zero, err
	}

	f, err := uc.folderRepo.GetFolder(ctx, input.FolderID)
	if err != nil {
		return zero, fmt.Errorf("loading folder: %w", err)
	}

	f.Name = input.Name
	f.UpdatedAt = uc.clock.Now()

	if err := uc.folderRepo.UpdateFolder(ctx, f); err != nil {
		return zero, fmt.Errorf("updating folder: %w", err)
	}

	return f, nil
}
