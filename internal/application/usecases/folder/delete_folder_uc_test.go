package folder

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"alp-drive/internal/domain/file"
)

type fakeDeleteFolderRepository struct {
	exists       bool
	existsErr    error
	cascadeFiles []file.File
	cascadeErr   error
}

func (f *fakeDeleteFolderRepository) ExistsFolderByID(_ context.Context, _ uuid.UUID) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeDeleteFolderRepository) DeleteFolderCascade(_ context.Context, _ uuid.UUID) ([]file.File, error) {
	return f.cascadeFiles, f.cascadeErr
}

type fakeFileDeleter struct {
	deletedIDs []uuid.UUID
	err        error
}

func (f *fakeFileDeleter) Delete(_ context.Context, fileID uuid.UUID) error {
	f.deletedIDs = append(f.deletedIDs, fileID)
	return f.err
}

func TestDeleteFolderUC_Execute(t *testing.T) {
	folderID := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b80")
	errBoom := errors.New("boom")

	t.Run("empty folder id", func(t *testing.T) {
		repo := &fakeDeleteFolderRepository{}
		deleter := &fakeFileDeleter{}
		uc := NewDeleteFolderUC(repo, deleter)

		err := uc.Execute(context.Background(), DeleteFolderInput{})
		if !errors.Is(err, ErrFolderIDEmpty) {
			t.Errorf("err = %v, want ErrFolderIDEmpty", err)
		}
	})

	t.Run("checking folder fails", func(t *testing.T) {
		repo := &fakeDeleteFolderRepository{existsErr: errBoom}
		deleter := &fakeFileDeleter{}
		uc := NewDeleteFolderUC(repo, deleter)

		err := uc.Execute(context.Background(), DeleteFolderInput{FolderID: folderID})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("folder not found", func(t *testing.T) {
		repo := &fakeDeleteFolderRepository{exists: false}
		deleter := &fakeFileDeleter{}
		uc := NewDeleteFolderUC(repo, deleter)

		err := uc.Execute(context.Background(), DeleteFolderInput{FolderID: folderID})
		if !errors.Is(err, ErrFolderNotFound) {
			t.Errorf("err = %v, want ErrFolderNotFound", err)
		}
	})

	t.Run("cascade delete fails", func(t *testing.T) {
		repo := &fakeDeleteFolderRepository{exists: true, cascadeErr: errBoom}
		deleter := &fakeFileDeleter{}
		uc := NewDeleteFolderUC(repo, deleter)

		err := uc.Execute(context.Background(), DeleteFolderInput{FolderID: folderID})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("deletes storage content for every deleted file, ignoring deleter errors", func(t *testing.T) {
		fileID1 := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b81")
		fileID2 := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b82")
		repo := &fakeDeleteFolderRepository{
			exists: true,
			cascadeFiles: []file.File{
				{ID: fileID1},
				{ID: fileID2},
			},
		}
		deleter := &fakeFileDeleter{err: errors.New("storage unreachable")}
		uc := NewDeleteFolderUC(repo, deleter)

		err := uc.Execute(context.Background(), DeleteFolderInput{FolderID: folderID})
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if len(deleter.deletedIDs) != 2 || deleter.deletedIDs[0] != fileID1 || deleter.deletedIDs[1] != fileID2 {
			t.Errorf("deletedIDs = %v, want [%v %v]", deleter.deletedIDs, fileID1, fileID2)
		}
	})
}
