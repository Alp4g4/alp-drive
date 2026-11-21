package file

import (
	"context"
	"fmt"
	"io"
	"strings"
	"uuid"

	"alp-drive/internal/application/ports/clock"
	"alp-drive/internal/domain/file"
)

type createFileRepository interface {
	ExistsFolderByID(ctx context.Context, id uuid.UUID) (bool, error)
	CreateFile(ctx context.Context, f file.File) error
}

type fileStorer interface {
	Store(ctx context.Context, fileID uuid.UUID, content io.Reader) error
}

type CreateFileUC struct {
	fileRepo   createFileRepository
	fileStorer fileStorer
	clock      clock.Clock
}

func NewCreateFileUC(fileRepo createFileRepository, storer fileStorer, clk clock.Clock) *CreateFileUC {
	return &CreateFileUC{
		fileRepo:   fileRepo,
		fileStorer: storer,
		clock:      clk,
	}
}

type CreateFileInput struct {
	Name     string
	FolderID uuid.UUID
	Content  io.Reader
	Size     int64
	MimeType string
}

func (i CreateFileInput) Validate() error {
	if strings.TrimSpace(i.Name) == "" {
		return ErrFileNameEmpty
	}

	if i.FolderID == uuid.Nil() {
		return ErrFolderIDEmpty
	}

	if i.Content == nil {
		return ErrFileContentNil
	}

	if i.Size < 0 {
		return ErrFileSizeNegative
	}

	if strings.TrimSpace(i.MimeType) == "" {
		return ErrFileMimeTypeEmpty
	}

	return nil
}

func (uc *CreateFileUC) Execute(ctx context.Context, input CreateFileInput) (file.File, error) {
	zero := file.File{}

	if err := input.Validate(); err != nil {
		return zero, err
	}

	exists, err := uc.fileRepo.ExistsFolderByID(ctx, input.FolderID)
	if err != nil {
		return zero, fmt.Errorf("checking target folder: %w", err)
	}

	if !exists {
		return zero, ErrFolderNotFound
	}

	now := uc.clock.Now()

	f := file.File{
		ID:        uuid.NewV7(),
		Name:      input.Name,
		Folder:    input.FolderID,
		Size:      input.Size,
		MimeType:  input.MimeType,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.fileStorer.Store(ctx, f.ID, input.Content); err != nil {
		return zero, fmt.Errorf("storing file: %w", err)
	}

	if err := uc.fileRepo.CreateFile(ctx, f); err != nil {
		return zero, fmt.Errorf("creating file: %w", err)
	}

	return f, nil
}
