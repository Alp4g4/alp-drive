package folder

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"alp-drive/internal/domain/folder"
)

type fakeListFoldersRepository struct {
	gotParentID *uuid.UUID
	folders     []folder.Folder
	err         error
}

func (f *fakeListFoldersRepository) ListFoldersByParentID(_ context.Context, parentID *uuid.UUID) ([]folder.Folder, error) {
	f.gotParentID = parentID
	return f.folders, f.err
}

func TestListFoldersUC_Execute(t *testing.T) {
	errBoom := errors.New("boom")

	t.Run("repository fails", func(t *testing.T) {
		repo := &fakeListFoldersRepository{err: errBoom}
		uc := NewListFoldersUC(repo)

		_, err := uc.Execute(context.Background(), ListFoldersInput{})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("returns folders", func(t *testing.T) {
		parentID := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b80")
		want := []folder.Folder{{Name: "docs"}, {Name: "photos"}}
		repo := &fakeListFoldersRepository{folders: want}
		uc := NewListFoldersUC(repo)

		got, err := uc.Execute(context.Background(), ListFoldersInput{ParentID: &parentID})
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if len(got) != len(want) {
			t.Fatalf("got %d folders, want %d", len(got), len(want))
		}
		if repo.gotParentID == nil || *repo.gotParentID != parentID {
			t.Errorf("repository received parentID %v, want %v", repo.gotParentID, parentID)
		}
	})
}
