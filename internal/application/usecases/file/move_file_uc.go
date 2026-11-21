package file

import (
	"context"
	"fmt"
	"uuid"

	"alp-drive/internal/application/ports/clock"
	"alp-drive/internal/domain/file"
)

type moveFileRepository interface {
	GetFile(ctx context.Context, id uuid.UUID) (file.File, error)
	ExistsFolderByID(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateFile(ctx context.Context, f file.File) error
}

type MoveFileUC struct {
	fileRepo moveFileRepository
	clock    clock.Clock
}

func NewMoveFileUC(fileRepo moveFileRepository, clk clock.Clock) *MoveFileUC {
	return &MoveFileUC{
		fileRepo: fileRepo,
		clock:    clk,
	}
}

type MoveFileInput struct {
	FileID      uuid.UUID
	NewFolderID uuid.UUID
}

func (i MoveFileInput) Validate() error {
	if i.FileID == uuid.Nil() {
		return ErrFileIDEmpty
	}

	if i.NewFolderID == uuid.Nil() {
		return ErrFolderIDEmpty
	}

	return nil
}

func (uc *MoveFileUC) Execute(ctx context.Context, input MoveFileInput) (file.File, error) {
	zero := file.File{}

	if err := input.Validate(); err != nil {
		return zero, err
	}

	f, err := uc.fileRepo.GetFile(ctx, input.FileID)
	if err != nil {
		return zero, fmt.Errorf("loading file: %w", err)
	}

	exists, err := uc.fileRepo.ExistsFolderByID(ctx, input.NewFolderID)
	if err != nil {
		return zero, fmt.Errorf("checking target folder: %w", err)
	}

	if !exists {
		return zero, ErrFolderNotFound
	}

	f.Folder = input.NewFolderID
	f.UpdatedAt = uc.clock.Now()

	if err := uc.fileRepo.UpdateFile(ctx, f); err != nil {
		return zero, fmt.Errorf("updating file: %w", err)
	}

	return f, nil
}
