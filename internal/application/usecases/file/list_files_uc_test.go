package file

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"alp-drive/internal/domain/file"
)

type fakeListFilesRepository struct {
	gotFolderID uuid.UUID
	files       []file.File
	err         error
}

func (f *fakeListFilesRepository) ListFilesByFolderID(_ context.Context, folderID uuid.UUID) ([]file.File, error) {
	f.gotFolderID = folderID
	return f.files, f.err
}

func TestListFilesByFolderUC_Execute(t *testing.T) {
	folderID := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b80")
	errBoom := errors.New("boom")

	t.Run("empty folder id", func(t *testing.T) {
		repo := &fakeListFilesRepository{}
		uc := NewListFilesByFolderUC(repo)

		_, err := uc.Execute(context.Background(), ListFilesByFolderInput{})
		if !errors.Is(err, ErrFolderIDEmpty) {
			t.Errorf("err = %v, want ErrFolderIDEmpty", err)
		}
	})

	t.Run("repository fails", func(t *testing.T) {
		repo := &fakeListFilesRepository{err: errBoom}
		uc := NewListFilesByFolderUC(repo)

		_, err := uc.Execute(context.Background(), ListFilesByFolderInput{FolderID: folderID})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("returns files", func(t *testing.T) {
		want := []file.File{{Name: "a.txt"}, {Name: "b.txt"}}
		repo := &fakeListFilesRepository{files: want}
		uc := NewListFilesByFolderUC(repo)

		got, err := uc.Execute(context.Background(), ListFilesByFolderInput{FolderID: folderID})
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if len(got) != len(want) {
			t.Fatalf("got %d files, want %d", len(got), len(want))
		}
		if repo.gotFolderID != folderID {
			t.Errorf("repository received folderID %v, want %v", repo.gotFolderID, folderID)
		}
	})
}
