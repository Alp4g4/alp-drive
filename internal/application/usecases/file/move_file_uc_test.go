package file

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"alp-drive/internal/domain/file"
)

type fakeMoveFileRepository struct {
	getFile   file.File
	getErr    error
	exists    bool
	existsErr error
	updateErr error
	updated   file.File
}

func (f *fakeMoveFileRepository) GetFile(_ context.Context, _ uuid.UUID) (file.File, error) {
	return f.getFile, f.getErr
}

func (f *fakeMoveFileRepository) ExistsFolderByID(_ context.Context, _ uuid.UUID) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeMoveFileRepository) UpdateFile(_ context.Context, fi file.File) error {
	f.updated = fi
	return f.updateErr
}

func TestMoveFileUC_Execute(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	fileID := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b80")
	newFolderID := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b81")
	errBoom := errors.New("boom")

	t.Run("empty file id", func(t *testing.T) {
		repo := &fakeMoveFileRepository{}
		uc := NewMoveFileUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), MoveFileInput{NewFolderID: newFolderID})
		if !errors.Is(err, ErrFileIDEmpty) {
			t.Errorf("err = %v, want ErrFileIDEmpty", err)
		}
	})

	t.Run("empty new folder id", func(t *testing.T) {
		repo := &fakeMoveFileRepository{}
		uc := NewMoveFileUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), MoveFileInput{FileID: fileID})
		if !errors.Is(err, ErrFolderIDEmpty) {
			t.Errorf("err = %v, want ErrFolderIDEmpty", err)
		}
	})

	t.Run("loading file fails", func(t *testing.T) {
		repo := &fakeMoveFileRepository{getErr: errBoom}
		uc := NewMoveFileUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), MoveFileInput{FileID: fileID, NewFolderID: newFolderID})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("checking target folder fails", func(t *testing.T) {
		repo := &fakeMoveFileRepository{existsErr: errBoom}
		uc := NewMoveFileUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), MoveFileInput{FileID: fileID, NewFolderID: newFolderID})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("target folder not found", func(t *testing.T) {
		repo := &fakeMoveFileRepository{exists: false}
		uc := NewMoveFileUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), MoveFileInput{FileID: fileID, NewFolderID: newFolderID})
		if !errors.Is(err, ErrFolderNotFound) {
			t.Errorf("err = %v, want ErrFolderNotFound", err)
		}
	})

	t.Run("updating file fails", func(t *testing.T) {
		repo := &fakeMoveFileRepository{exists: true, updateErr: errBoom}
		uc := NewMoveFileUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), MoveFileInput{FileID: fileID, NewFolderID: newFolderID})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("moved", func(t *testing.T) {
		repo := &fakeMoveFileRepository{exists: true, getFile: file.File{ID: fileID}}
		uc := NewMoveFileUC(repo, fixedClock{now: now})

		f, err := uc.Execute(context.Background(), MoveFileInput{FileID: fileID, NewFolderID: newFolderID})
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if f.Folder != newFolderID {
			t.Errorf("Folder = %v, want %v", f.Folder, newFolderID)
		}
		if !f.UpdatedAt.Equal(now) {
			t.Errorf("UpdatedAt = %v, want %v", f.UpdatedAt, now)
		}
	})
}
