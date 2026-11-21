package file

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"alp-drive/internal/domain/file"
)

type fakeRenameFileRepository struct {
	getFile   file.File
	getErr    error
	updateErr error
	updated   file.File
}

func (f *fakeRenameFileRepository) GetFile(_ context.Context, _ uuid.UUID) (file.File, error) {
	return f.getFile, f.getErr
}

func (f *fakeRenameFileRepository) UpdateFile(_ context.Context, fi file.File) error {
	f.updated = fi
	return f.updateErr
}

func TestRenameFileUC_Execute(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	fileID := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b80")
	errBoom := errors.New("boom")

	t.Run("empty file id", func(t *testing.T) {
		repo := &fakeRenameFileRepository{}
		uc := NewRenameFileUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), RenameFileInput{Name: "new.txt"})
		if !errors.Is(err, ErrFileIDEmpty) {
			t.Errorf("err = %v, want ErrFileIDEmpty", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		repo := &fakeRenameFileRepository{}
		uc := NewRenameFileUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), RenameFileInput{FileID: fileID, Name: "  "})
		if !errors.Is(err, ErrFileNameEmpty) {
			t.Errorf("err = %v, want ErrFileNameEmpty", err)
		}
	})

	t.Run("loading file fails", func(t *testing.T) {
		repo := &fakeRenameFileRepository{getErr: errBoom}
		uc := NewRenameFileUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), RenameFileInput{FileID: fileID, Name: "new.txt"})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("updating file fails", func(t *testing.T) {
		repo := &fakeRenameFileRepository{updateErr: errBoom}
		uc := NewRenameFileUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), RenameFileInput{FileID: fileID, Name: "new.txt"})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("renamed", func(t *testing.T) {
		repo := &fakeRenameFileRepository{getFile: file.File{ID: fileID, Name: "old.txt"}}
		uc := NewRenameFileUC(repo, fixedClock{now: now})

		f, err := uc.Execute(context.Background(), RenameFileInput{FileID: fileID, Name: "new.txt"})
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if f.Name != "new.txt" {
			t.Errorf("Name = %q, want %q", f.Name, "new.txt")
		}
		if !f.UpdatedAt.Equal(now) {
			t.Errorf("UpdatedAt = %v, want %v", f.UpdatedAt, now)
		}
		if repo.updated.Name != "new.txt" {
			t.Errorf("repository received Name = %q, want %q", repo.updated.Name, "new.txt")
		}
	})
}
