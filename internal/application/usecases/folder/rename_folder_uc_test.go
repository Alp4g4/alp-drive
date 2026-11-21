package folder

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"alp-drive/internal/domain/folder"
)

type fakeRenameFolderRepository struct {
	getFolder folder.Folder
	getErr    error
	updateErr error
	updated   folder.Folder
}

func (f *fakeRenameFolderRepository) GetFolder(_ context.Context, _ uuid.UUID) (folder.Folder, error) {
	return f.getFolder, f.getErr
}

func (f *fakeRenameFolderRepository) UpdateFolder(_ context.Context, fo folder.Folder) error {
	f.updated = fo
	return f.updateErr
}

func TestRenameFolderUC_Execute(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	folderID := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b80")
	errBoom := errors.New("boom")

	t.Run("empty folder id", func(t *testing.T) {
		repo := &fakeRenameFolderRepository{}
		uc := NewRenameFolderUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), RenameFolderInput{Name: "docs"})
		if !errors.Is(err, ErrFolderIDEmpty) {
			t.Errorf("err = %v, want ErrFolderIDEmpty", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		repo := &fakeRenameFolderRepository{}
		uc := NewRenameFolderUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), RenameFolderInput{FolderID: folderID, Name: "  "})
		if !errors.Is(err, ErrFolderNameEmpty) {
			t.Errorf("err = %v, want ErrFolderNameEmpty", err)
		}
	})

	t.Run("loading folder fails", func(t *testing.T) {
		repo := &fakeRenameFolderRepository{getErr: errBoom}
		uc := NewRenameFolderUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), RenameFolderInput{FolderID: folderID, Name: "docs"})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("updating folder fails", func(t *testing.T) {
		repo := &fakeRenameFolderRepository{updateErr: errBoom}
		uc := NewRenameFolderUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), RenameFolderInput{FolderID: folderID, Name: "docs"})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("renamed", func(t *testing.T) {
		repo := &fakeRenameFolderRepository{getFolder: folder.Folder{ID: folderID, Name: "old"}}
		uc := NewRenameFolderUC(repo, fixedClock{now: now})

		f, err := uc.Execute(context.Background(), RenameFolderInput{FolderID: folderID, Name: "new"})
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if f.Name != "new" {
			t.Errorf("Name = %q, want %q", f.Name, "new")
		}
		if !f.UpdatedAt.Equal(now) {
			t.Errorf("UpdatedAt = %v, want %v", f.UpdatedAt, now)
		}
		if repo.updated.Name != "new" {
			t.Errorf("repository received Name = %q, want %q", repo.updated.Name, "new")
		}
	})
}
