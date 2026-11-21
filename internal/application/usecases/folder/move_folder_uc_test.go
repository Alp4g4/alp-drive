package folder

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"alp-drive/internal/domain/folder"
)

type fakeMoveFolderRepository struct {
	getFolder folder.Folder
	getErr    error
	exists    bool
	existsErr error
	updateErr error
	updated   folder.Folder
}

func (f *fakeMoveFolderRepository) GetFolder(_ context.Context, _ uuid.UUID) (folder.Folder, error) {
	return f.getFolder, f.getErr
}

func (f *fakeMoveFolderRepository) ExistsFolderByID(_ context.Context, _ uuid.UUID) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeMoveFolderRepository) UpdateFolder(_ context.Context, fo folder.Folder) error {
	f.updated = fo
	return f.updateErr
}

func TestMoveFolderUC_Execute(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	folderID := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b80")
	otherParentID := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b81")
	errBoom := errors.New("boom")

	t.Run("empty folder id", func(t *testing.T) {
		repo := &fakeMoveFolderRepository{}
		uc := NewMoveFolderUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), MoveFolderInput{})
		if !errors.Is(err, ErrFolderIDEmpty) {
			t.Errorf("err = %v, want ErrFolderIDEmpty", err)
		}
	})

	t.Run("loading folder fails", func(t *testing.T) {
		repo := &fakeMoveFolderRepository{getErr: errBoom}
		uc := NewMoveFolderUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), MoveFolderInput{FolderID: folderID})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("checking new parent fails", func(t *testing.T) {
		repo := &fakeMoveFolderRepository{
			getFolder: folder.Folder{ID: folderID},
			existsErr: errBoom,
		}
		uc := NewMoveFolderUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), MoveFolderInput{FolderID: folderID, NewParentID: &otherParentID})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("new parent not found", func(t *testing.T) {
		repo := &fakeMoveFolderRepository{
			getFolder: folder.Folder{ID: folderID},
			exists:    false,
		}
		uc := NewMoveFolderUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), MoveFolderInput{FolderID: folderID, NewParentID: &otherParentID})
		if !errors.Is(err, ErrParentFolderNotFound) {
			t.Errorf("err = %v, want ErrParentFolderNotFound", err)
		}
	})

	t.Run("folder cannot be its own parent", func(t *testing.T) {
		repo := &fakeMoveFolderRepository{
			getFolder: folder.Folder{ID: folderID},
			exists:    true,
		}
		uc := NewMoveFolderUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), MoveFolderInput{FolderID: folderID, NewParentID: &folderID})
		if !errors.Is(err, folder.ErrFolderCannotBeOwnParent) {
			t.Errorf("err = %v, want ErrFolderCannotBeOwnParent", err)
		}
	})

	t.Run("updating folder fails", func(t *testing.T) {
		repo := &fakeMoveFolderRepository{
			getFolder: folder.Folder{ID: folderID},
			updateErr: errBoom,
		}
		uc := NewMoveFolderUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), MoveFolderInput{FolderID: folderID})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("moved to root", func(t *testing.T) {
		repo := &fakeMoveFolderRepository{
			getFolder: folder.Folder{ID: folderID, Parent: &otherParentID},
		}
		uc := NewMoveFolderUC(repo, fixedClock{now: now})

		f, err := uc.Execute(context.Background(), MoveFolderInput{FolderID: folderID})
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if f.Parent != nil {
			t.Errorf("Parent = %v, want nil", f.Parent)
		}
		if !f.UpdatedAt.Equal(now) {
			t.Errorf("UpdatedAt = %v, want %v", f.UpdatedAt, now)
		}
	})

	t.Run("moved under a new parent", func(t *testing.T) {
		repo := &fakeMoveFolderRepository{
			getFolder: folder.Folder{ID: folderID},
			exists:    true,
		}
		uc := NewMoveFolderUC(repo, fixedClock{now: now})

		f, err := uc.Execute(context.Background(), MoveFolderInput{FolderID: folderID, NewParentID: &otherParentID})
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if f.Parent == nil || *f.Parent != otherParentID {
			t.Errorf("Parent = %v, want %v", f.Parent, otherParentID)
		}
	})
}
