package file

import (
	"context"
	"fmt"
	"strings"
	"uuid"

	"alp-drive/internal/application/ports/clock"
	"alp-drive/internal/domain/file"
)

type renameFileRepository interface {
	GetFile(ctx context.Context, id uuid.UUID) (file.File, error)
	UpdateFile(ctx context.Context, f file.File) error
}

type RenameFileUC struct {
	fileRepo renameFileRepository
	clock    clock.Clock
}

func NewRenameFileUC(fileRepo renameFileRepository, clk clock.Clock) *RenameFileUC {
	return &RenameFileUC{
		fileRepo: fileRepo,
		clock:    clk,
	}
}

type RenameFileInput struct {
	FileID uuid.UUID
	Name   string
}

func (i RenameFileInput) Validate() error {
	if i.FileID == uuid.Nil() {
		return ErrFileIDEmpty
	}

	if strings.TrimSpace(i.Name) == "" {
		return ErrFileNameEmpty
	}

	return nil
}

func (uc *RenameFileUC) Execute(ctx context.Context, input RenameFileInput) (file.File, error) {
	zero := file.File{}

	if err := input.Validate(); err != nil {
		return zero, err
	}

	f, err := uc.fileRepo.GetFile(ctx, input.FileID)
	if err != nil {
		return zero, fmt.Errorf("loading file: %w", err)
	}

	f.Name = input.Name
	f.UpdatedAt = uc.clock.Now()

	if err := uc.fileRepo.UpdateFile(ctx, f); err != nil {
		return zero, fmt.Errorf("updating file: %w", err)
	}

	return f, nil
}
