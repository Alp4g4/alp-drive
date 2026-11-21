package folder

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"alp-drive/internal/domain/folder"
)

type fakeCreateFolderRepository struct {
	exists    bool
	existsErr error
	createErr error
	created   folder.Folder
}

func (f *fakeCreateFolderRepository) ExistsFolderByID(_ context.Context, _ uuid.UUID) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeCreateFolderRepository) CreateFolder(_ context.Context, fo folder.Folder) error {
	f.created = fo
	return f.createErr
}

func TestCreateFolderUC_Execute(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	parentID := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b80")
	errBoom := errors.New("boom")

	t.Run("empty name", func(t *testing.T) {
		repo := &fakeCreateFolderRepository{}
		uc := NewCreateFolderUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), CreateFolderInput{Name: "   "})
		if !errors.Is(err, ErrFolderNameEmpty) {
			t.Errorf("err = %v, want ErrFolderNameEmpty", err)
		}
	})

	t.Run("checking parent folder fails", func(t *testing.T) {
		repo := &fakeCreateFolderRepository{existsErr: errBoom}
		uc := NewCreateFolderUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), CreateFolderInput{Name: "docs", ParentID: &parentID})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("parent folder not found", func(t *testing.T) {
		repo := &fakeCreateFolderRepository{exists: false}
		uc := NewCreateFolderUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), CreateFolderInput{Name: "docs", ParentID: &parentID})
		if !errors.Is(err, ErrParentFolderNotFound) {
			t.Errorf("err = %v, want ErrParentFolderNotFound", err)
		}
	})

	t.Run("create fails", func(t *testing.T) {
		repo := &fakeCreateFolderRepository{createErr: errBoom}
		uc := NewCreateFolderUC(repo, fixedClock{now: now})

		_, err := uc.Execute(context.Background(), CreateFolderInput{Name: "docs"})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("created at root", func(t *testing.T) {
		repo := &fakeCreateFolderRepository{}
		uc := NewCreateFolderUC(repo, fixedClock{now: now})

		f, err := uc.Execute(context.Background(), CreateFolderInput{Name: "docs"})
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if f.ID == uuid.Nil() {
			t.Error("ID was not generated")
		}
		if f.Name != "docs" {
			t.Errorf("Name = %q, want %q", f.Name, "docs")
		}
		if f.Parent != nil {
			t.Errorf("Parent = %v, want nil", f.Parent)
		}
		if !f.CreatedAt.Equal(now) || !f.UpdatedAt.Equal(now) {
			t.Errorf("CreatedAt/UpdatedAt = %v/%v, want %v", f.CreatedAt, f.UpdatedAt, now)
		}
		if repo.created.ID != f.ID {
			t.Errorf("repository received a different folder than the one returned")
		}
	})

	t.Run("created under a parent", func(t *testing.T) {
		repo := &fakeCreateFolderRepository{exists: true}
		uc := NewCreateFolderUC(repo, fixedClock{now: now})

		f, err := uc.Execute(context.Background(), CreateFolderInput{Name: "docs", ParentID: &parentID})
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if f.Parent == nil || *f.Parent != parentID {
			t.Errorf("Parent = %v, want %v", f.Parent, parentID)
		}
	})
}
