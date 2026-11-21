package file

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"
	"uuid"

	"alp-drive/internal/domain/file"
)

type fakeCreateFileRepository struct {
	exists    bool
	existsErr error
	createErr error
	created   file.File
}

func (f *fakeCreateFileRepository) ExistsFolderByID(_ context.Context, _ uuid.UUID) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeCreateFileRepository) CreateFile(_ context.Context, fi file.File) error {
	f.created = fi
	return f.createErr
}

type fakeFileStorer struct {
	err        error
	gotFileID  uuid.UUID
	gotContent []byte
}

func (s *fakeFileStorer) Store(_ context.Context, fileID uuid.UUID, content io.Reader) error {
	s.gotFileID = fileID
	b, _ := io.ReadAll(content)
	s.gotContent = b
	return s.err
}

func validCreateFileInput() CreateFileInput {
	folderID := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b80")
	return CreateFileInput{
		Name:     "report.pdf",
		FolderID: folderID,
		Content:  bytes.NewReader([]byte("hello world")),
		Size:     11,
		MimeType: "application/pdf",
	}
}

func TestCreateFileUC_Execute(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	errBoom := errors.New("boom")

	t.Run("empty name", func(t *testing.T) {
		repo := &fakeCreateFileRepository{}
		storer := &fakeFileStorer{}
		uc := NewCreateFileUC(repo, storer, fixedClock{now: now})

		input := validCreateFileInput()
		input.Name = "  "
		_, err := uc.Execute(context.Background(), input)
		if !errors.Is(err, ErrFileNameEmpty) {
			t.Errorf("err = %v, want ErrFileNameEmpty", err)
		}
	})

	t.Run("empty folder id", func(t *testing.T) {
		repo := &fakeCreateFileRepository{}
		storer := &fakeFileStorer{}
		uc := NewCreateFileUC(repo, storer, fixedClock{now: now})

		input := validCreateFileInput()
		input.FolderID = uuid.Nil()
		_, err := uc.Execute(context.Background(), input)
		if !errors.Is(err, ErrFolderIDEmpty) {
			t.Errorf("err = %v, want ErrFolderIDEmpty", err)
		}
	})

	t.Run("nil content", func(t *testing.T) {
		repo := &fakeCreateFileRepository{}
		storer := &fakeFileStorer{}
		uc := NewCreateFileUC(repo, storer, fixedClock{now: now})

		input := validCreateFileInput()
		input.Content = nil
		_, err := uc.Execute(context.Background(), input)
		if !errors.Is(err, ErrFileContentNil) {
			t.Errorf("err = %v, want ErrFileContentNil", err)
		}
	})

	t.Run("negative size", func(t *testing.T) {
		repo := &fakeCreateFileRepository{}
		storer := &fakeFileStorer{}
		uc := NewCreateFileUC(repo, storer, fixedClock{now: now})

		input := validCreateFileInput()
		input.Size = -1
		_, err := uc.Execute(context.Background(), input)
		if !errors.Is(err, ErrFileSizeNegative) {
			t.Errorf("err = %v, want ErrFileSizeNegative", err)
		}
	})

	t.Run("empty mime type", func(t *testing.T) {
		repo := &fakeCreateFileRepository{}
		storer := &fakeFileStorer{}
		uc := NewCreateFileUC(repo, storer, fixedClock{now: now})

		input := validCreateFileInput()
		input.MimeType = "  "
		_, err := uc.Execute(context.Background(), input)
		if !errors.Is(err, ErrFileMimeTypeEmpty) {
			t.Errorf("err = %v, want ErrFileMimeTypeEmpty", err)
		}
	})

	t.Run("checking target folder fails", func(t *testing.T) {
		repo := &fakeCreateFileRepository{existsErr: errBoom}
		storer := &fakeFileStorer{}
		uc := NewCreateFileUC(repo, storer, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), validCreateFileInput())
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("target folder not found", func(t *testing.T) {
		repo := &fakeCreateFileRepository{exists: false}
		storer := &fakeFileStorer{}
		uc := NewCreateFileUC(repo, storer, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), validCreateFileInput())
		if !errors.Is(err, ErrFolderNotFound) {
			t.Errorf("err = %v, want ErrFolderNotFound", err)
		}
	})

	t.Run("storing file fails", func(t *testing.T) {
		repo := &fakeCreateFileRepository{exists: true}
		storer := &fakeFileStorer{err: errBoom}
		uc := NewCreateFileUC(repo, storer, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), validCreateFileInput())
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("creating file fails", func(t *testing.T) {
		repo := &fakeCreateFileRepository{exists: true, createErr: errBoom}
		storer := &fakeFileStorer{}
		uc := NewCreateFileUC(repo, storer, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), validCreateFileInput())
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("created", func(t *testing.T) {
		repo := &fakeCreateFileRepository{exists: true}
		storer := &fakeFileStorer{}
		uc := NewCreateFileUC(repo, storer, fixedClock{now: now})

		input := validCreateFileInput()
		f, err := uc.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if f.ID == uuid.Nil() {
			t.Error("ID was not generated")
		}
		if f.Name != input.Name || f.Folder != input.FolderID || f.Size != input.Size || f.MimeType != input.MimeType {
			t.Errorf("f = %+v, want fields from input %+v", f, input)
		}
		if !f.CreatedAt.Equal(now) || !f.UpdatedAt.Equal(now) {
			t.Errorf("CreatedAt/UpdatedAt = %v/%v, want %v", f.CreatedAt, f.UpdatedAt, now)
		}
		if storer.gotFileID != f.ID {
			t.Errorf("storer received fileID %v, want %v", storer.gotFileID, f.ID)
		}
		if string(storer.gotContent) != "hello world" {
			t.Errorf("storer received content %q, want %q", storer.gotContent, "hello world")
		}
		if repo.created.ID != f.ID {
			t.Errorf("repository received ID = %v, want %v", repo.created.ID, f.ID)
		}
	})
}
