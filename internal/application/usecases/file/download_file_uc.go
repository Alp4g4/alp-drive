package file

import (
	"context"
	"fmt"
	"io"
	"uuid"

	"alp-drive/internal/domain/file"
)

type downloadFileRepository interface {
	GetFile(ctx context.Context, id uuid.UUID) (file.File, error)
}

type fileReader interface {
	Open(ctx context.Context, fileID uuid.UUID) (io.ReadCloser, error)
}

type DownloadFileUC struct {
	fileRepo   downloadFileRepository
	fileReader fileReader
}

func NewDownloadFileUC(fileRepo downloadFileRepository, fileReader fileReader) *DownloadFileUC {
	return &DownloadFileUC{
		fileRepo:   fileRepo,
		fileReader: fileReader,
	}
}

type DownloadFileInput struct {
	FileID uuid.UUID
}

func (i DownloadFileInput) Validate() error {
	if i.FileID == uuid.Nil() {
		return ErrFileIDEmpty
	}

	return nil
}

type DownloadFileOutput struct {
	File    file.File
	Content io.ReadCloser
}

func (uc *DownloadFileUC) Execute(ctx context.Context, input DownloadFileInput) (DownloadFileOutput, error) {
	zero := DownloadFileOutput{}

	if err := input.Validate(); err != nil {
		return zero, err
	}

	f, err := uc.fileRepo.GetFile(ctx, input.FileID)
	if err != nil {
		return zero, fmt.Errorf("loading file: %w", err)
	}

	content, err := uc.fileReader.Open(ctx, f.ID)
	if err != nil {
		return zero, fmt.Errorf("opening file content: %w", err)
	}

	return DownloadFileOutput{
		File:    f,
		Content: content,
	}, nil
}
