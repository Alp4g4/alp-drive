package file

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"alp-drive/internal/domain/file"
)

type fakeDeleteFileRepository struct {
	getFile   file.File
	getErr    error
	deleteErr error
	deletedID uuid.UUID
}

func (f *fakeDeleteFileRepository) GetFile(_ context.Context, _ uuid.UUID) (file.File, error) {
	return f.getFile, f.getErr
}

func (f *fakeDeleteFileRepository) DeleteFile(_ context.Context, id uuid.UUID) error {
	f.deletedID = id
	return f.deleteErr
}

type fakeFileDeleter struct {
	gotID uuid.UUID
	err   error
}

func (f *fakeFileDeleter) Delete(_ context.Context, fileID uuid.UUID) error {
	f.gotID = fileID
	return f.err
}

func TestDeleteFileUC_Execute(t *testing.T) {
	fileID := uuid.MustParse("019a8f2c-6d41-7b3e-9a57-3f1e2d4c6b80")
	errBoom := errors.New("boom")

	t.Run("empty file id", func(t *testing.T) {
		repo := &fakeDeleteFileRepository{}
		deleter := &fakeFileDeleter{}
		uc := NewDeleteFileUC(repo, deleter)

		err := uc.Execute(context.Background(), DeleteFileInput{})
		if !errors.Is(err, ErrFileIDEmpty) {
			t.Errorf("err = %v, want ErrFileIDEmpty", err)
		}
	})

	t.Run("loading file fails", func(t *testing.T) {
		repo := &fakeDeleteFileRepository{getErr: errBoom}
		deleter := &fakeFileDeleter{}
		uc := NewDeleteFileUC(repo, deleter)

		err := uc.Execute(context.Background(), DeleteFileInput{FileID: fileID})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("deleting file fails", func(t *testing.T) {
		repo := &fakeDeleteFileRepository{deleteErr: errBoom}
		deleter := &fakeFileDeleter{}
		uc := NewDeleteFileUC(repo, deleter)

		err := uc.Execute(context.Background(), DeleteFileInput{FileID: fileID})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want wrapped errBoom", err)
		}
	})

	t.Run("deletes storage content, ignoring deleter errors", func(t *testing.T) {
		repo := &fakeDeleteFileRepository{getFile: file.File{ID: fileID}}
		deleter := &fakeFileDeleter{err: errors.New("storage unreachable")}
		uc := NewDeleteFileUC(repo, deleter)

		err := uc.Execute(context.Background(), DeleteFileInput{FileID: fileID})
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if repo.deletedID != fileID {
			t.Errorf("deletedID = %v, want %v", repo.deletedID, fileID)
		}
		if deleter.gotID != fileID {
			t.Errorf("deleter received id %v, want %v", deleter.gotID, fileID)
		}
	})
}
