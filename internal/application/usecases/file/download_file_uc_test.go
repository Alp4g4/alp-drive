package file

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"uuid"

	"alp-drive/internal/domain/file"
)

type fakeDownloadFileRepository struct {
	getFile file.File
	getErr  error
}

func (f *fakeDownloadFileRepository) GetFile(_ context.Context, _ uuid.UUID) (file.File, error) {
	return f.getFile, f.getErr
}

type fakeFileReader struct {
	content io.ReadCloser
	err     error
	gotID   uuid.UUID
}

func (f *fakeFileReader) Open(_ context.Context, fileID uuid.UUID) (io.ReadCloser, error) {
	f.gotID = fileID
	return f.content, f.err
}

func TestDownloadFileUC_Execute(t *testing.T) {
	fileID := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b80")
	errBoom := errors.New("boom")

	t.Run("empty file id", func(t *testing.T) {
		repo := &fakeDownloadFileRepository{}
		reader := &fakeFileReader{}
		uc := NewDownloadFileUC(repo, reader)

		_, err := uc.Execute(context.Background(), DownloadFileInput{})
		if !errors.Is(err, ErrFileIDEmpty) {
			t.Errorf("err = %v, want ErrFileIDEmpty", err)
		}
	})

	t.Run("loading file fails", func(t *testing.T) {
		repo := &fakeDownloadFileRepository{getErr: errBoom}
		reader := &fakeFileReader{}
		uc := NewDownloadFileUC(repo, reader)

		_, err := uc.Execute(context.Background(), DownloadFileInput{FileID: fileID})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("opening content fails", func(t *testing.T) {
		repo := &fakeDownloadFileRepository{getFile: file.File{ID: fileID}}
		reader := &fakeFileReader{err: errBoom}
		uc := NewDownloadFileUC(repo, reader)

		_, err := uc.Execute(context.Background(), DownloadFileInput{FileID: fileID})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("downloaded", func(t *testing.T) {
		want := file.File{ID: fileID, Name: "report.pdf"}
		repo := &fakeDownloadFileRepository{getFile: want}
		reader := &fakeFileReader{content: io.NopCloser(strings.NewReader("hello world"))}
		uc := NewDownloadFileUC(repo, reader)

		out, err := uc.Execute(context.Background(), DownloadFileInput{FileID: fileID})
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if out.File != want {
			t.Errorf("File = %+v, want %+v", out.File, want)
		}
		if reader.gotID != fileID {
			t.Errorf("reader received id %v, want %v", reader.gotID, fileID)
		}

		b, err := io.ReadAll(out.Content)
		if err != nil {
			t.Fatalf("reading content: %v", err)
		}
		if string(b) != "hello world" {
			t.Errorf("content = %q, want %q", b, "hello world")
		}
	})
}
