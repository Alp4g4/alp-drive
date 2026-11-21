package sqlite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
	"uuid"

	"github.com/pressly/goose/v3"

	usecasefile "alp-drive/internal/application/usecases/file"
	usecasefolder "alp-drive/internal/application/usecases/folder"
	"alp-drive/internal/domain/file"
	"alp-drive/internal/domain/folder"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	ctx := context.Background()

	db, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	provider, err := goose.NewProvider(goose.DialectSQLite3, db, os.DirFS("migrations"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatalf("migrating: %v", err)
	}

	return NewStore(db)
}

func newFolder(name string, parent *uuid.UUID) folder.Folder {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return folder.Folder{ID: uuid.NewV7(), Name: name, Parent: parent, CreatedAt: now, UpdatedAt: now}
}

func newFile(name string, folderID uuid.UUID) file.File {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return file.File{ID: uuid.NewV7(), Name: name, Folder: folderID, Size: 11, MimeType: "text/plain", CreatedAt: now, UpdatedAt: now}
}

func TestCreateAndGetFolder(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	f := newFolder("root", nil)

	if err := store.CreateFolder(ctx, f); err != nil {
		t.Fatalf("CreateFolder: %v", err)
	}

	got, err := store.GetFolder(ctx, f.ID)
	if err != nil {
		t.Fatalf("GetFolder: %v", err)
	}
	if got.ID != f.ID || got.Name != f.Name || got.Parent != nil || !got.CreatedAt.Equal(f.CreatedAt) {
		t.Errorf("GetFolder = %+v, want %+v", got, f)
	}

	exists, err := store.ExistsFolderByID(ctx, f.ID)
	if err != nil || !exists {
		t.Errorf("ExistsFolderByID = %v, %v, want true", exists, err)
	}
}

func TestGetMissingFolder(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if _, err := store.GetFolder(ctx, uuid.NewV7()); !errors.Is(err, usecasefolder.ErrFolderNotFound) {
		t.Errorf("GetFolder = %v, want ErrFolderNotFound", err)
	}
}

func TestUpdateFolder(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	root := newFolder("root", nil)
	if err := store.CreateFolder(ctx, root); err != nil {
		t.Fatal(err)
	}
	child := newFolder("child", &root.ID)
	if err := store.CreateFolder(ctx, child); err != nil {
		t.Fatal(err)
	}

	child.Name, child.Parent, child.UpdatedAt = "renamed", nil, child.UpdatedAt.Add(time.Minute)
	if err := store.UpdateFolder(ctx, child); err != nil {
		t.Fatalf("UpdateFolder: %v", err)
	}

	got, err := store.GetFolder(ctx, child.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "renamed" || got.Parent != nil || !got.UpdatedAt.Equal(child.UpdatedAt) {
		t.Errorf("after UpdateFolder = %+v, want %+v", got, child)
	}
}

func TestUpdateMissingFolder(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if err := store.UpdateFolder(ctx, newFolder("ghost", nil)); !errors.Is(err, usecasefolder.ErrFolderNotFound) {
		t.Errorf("UpdateFolder = %v, want ErrFolderNotFound", err)
	}
}

func TestListFoldersByParentID(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	root := newFolder("root", nil)
	if err := store.CreateFolder(ctx, root); err != nil {
		t.Fatal(err)
	}
	other := newFolder("other-root", nil)
	if err := store.CreateFolder(ctx, other); err != nil {
		t.Fatal(err)
	}
	childA := newFolder("a", &root.ID)
	childB := newFolder("b", &root.ID)
	if err := store.CreateFolder(ctx, childA); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateFolder(ctx, childB); err != nil {
		t.Fatal(err)
	}

	roots, err := store.ListFoldersByParentID(ctx, nil)
	if err != nil {
		t.Fatalf("ListFoldersByParentID(nil): %v", err)
	}
	if len(roots) != 2 {
		t.Fatalf("got %d root folders, want 2", len(roots))
	}

	children, err := store.ListFoldersByParentID(ctx, &root.ID)
	if err != nil {
		t.Fatalf("ListFoldersByParentID(root): %v", err)
	}
	if len(children) != 2 || children[0].Name != "a" || children[1].Name != "b" {
		t.Errorf("children = %+v, want [a b]", children)
	}
}

func TestCreateAndGetFile(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	root := newFolder("root", nil)
	if err := store.CreateFolder(ctx, root); err != nil {
		t.Fatal(err)
	}
	f := newFile("report.pdf", root.ID)

	if err := store.CreateFile(ctx, f); err != nil {
		t.Fatalf("CreateFile: %v", err)
	}

	got, err := store.GetFile(ctx, f.ID)
	if err != nil {
		t.Fatalf("GetFile: %v", err)
	}
	if got.ID != f.ID || got.Name != f.Name || got.Folder != f.Folder || got.Size != f.Size || got.MimeType != f.MimeType {
		t.Errorf("GetFile = %+v, want %+v", got, f)
	}
}

func TestGetMissingFile(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if _, err := store.GetFile(ctx, uuid.NewV7()); !errors.Is(err, usecasefile.ErrFileNotFound) {
		t.Errorf("GetFile = %v, want ErrFileNotFound", err)
	}
}

func TestUpdateFile(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	folderA := newFolder("a", nil)
	folderB := newFolder("b", nil)
	if err := store.CreateFolder(ctx, folderA); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateFolder(ctx, folderB); err != nil {
		t.Fatal(err)
	}
	f := newFile("report.pdf", folderA.ID)
	if err := store.CreateFile(ctx, f); err != nil {
		t.Fatal(err)
	}

	f.Folder, f.Name, f.UpdatedAt = folderB.ID, "renamed.pdf", f.UpdatedAt.Add(time.Minute)
	if err := store.UpdateFile(ctx, f); err != nil {
		t.Fatalf("UpdateFile: %v", err)
	}

	got, err := store.GetFile(ctx, f.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Folder != folderB.ID || got.Name != "renamed.pdf" {
		t.Errorf("after UpdateFile = %+v, want folder %v, name %q", got, folderB.ID, "renamed.pdf")
	}
}

func TestUpdateMissingFile(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if err := store.UpdateFile(ctx, newFile("ghost.pdf", uuid.NewV7())); !errors.Is(err, usecasefile.ErrFileNotFound) {
		t.Errorf("UpdateFile = %v, want ErrFileNotFound", err)
	}
}

func TestDeleteFile(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	root := newFolder("root", nil)
	if err := store.CreateFolder(ctx, root); err != nil {
		t.Fatal(err)
	}
	f := newFile("report.pdf", root.ID)
	if err := store.CreateFile(ctx, f); err != nil {
		t.Fatal(err)
	}

	if err := store.DeleteFile(ctx, f.ID); err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}

	if _, err := store.GetFile(ctx, f.ID); !errors.Is(err, usecasefile.ErrFileNotFound) {
		t.Errorf("GetFile after delete = %v, want ErrFileNotFound", err)
	}
}

func TestDeleteMissingFile(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if err := store.DeleteFile(ctx, uuid.NewV7()); !errors.Is(err, usecasefile.ErrFileNotFound) {
		t.Errorf("DeleteFile = %v, want ErrFileNotFound", err)
	}
}

func TestListFilesByFolderID(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	root := newFolder("root", nil)
	if err := store.CreateFolder(ctx, root); err != nil {
		t.Fatal(err)
	}
	other := newFolder("other", nil)
	if err := store.CreateFolder(ctx, other); err != nil {
		t.Fatal(err)
	}
	a := newFile("a.txt", root.ID)
	b := newFile("b.txt", root.ID)
	outside := newFile("outside.txt", other.ID)
	for _, f := range []file.File{a, b, outside} {
		if err := store.CreateFile(ctx, f); err != nil {
			t.Fatal(err)
		}
	}

	files, err := store.ListFilesByFolderID(ctx, root.ID)
	if err != nil {
		t.Fatalf("ListFilesByFolderID: %v", err)
	}
	if len(files) != 2 || files[0].Name != "a.txt" || files[1].Name != "b.txt" {
		t.Errorf("files = %+v, want [a.txt b.txt]", files)
	}
}

func TestDeleteFolderCascade(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	root := newFolder("root", nil)
	if err := store.CreateFolder(ctx, root); err != nil {
		t.Fatal(err)
	}
	child := newFolder("child", &root.ID)
	if err := store.CreateFolder(ctx, child); err != nil {
		t.Fatal(err)
	}
	rootFile := newFile("root.txt", root.ID)
	childFile := newFile("child.txt", child.ID)
	if err := store.CreateFile(ctx, rootFile); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateFile(ctx, childFile); err != nil {
		t.Fatal(err)
	}

	deleted, err := store.DeleteFolderCascade(ctx, root.ID)
	if err != nil {
		t.Fatalf("DeleteFolderCascade: %v", err)
	}
	if len(deleted) != 2 {
		t.Fatalf("got %d deleted files, want 2", len(deleted))
	}

	if exists, err := store.ExistsFolderByID(ctx, root.ID); err != nil || exists {
		t.Errorf("root still exists: %v, %v", exists, err)
	}
	if exists, err := store.ExistsFolderByID(ctx, child.ID); err != nil || exists {
		t.Errorf("child still exists: %v, %v", exists, err)
	}
	if _, err := store.GetFile(ctx, rootFile.ID); !errors.Is(err, usecasefile.ErrFileNotFound) {
		t.Errorf("rootFile GetFile = %v, want ErrFileNotFound", err)
	}
	if _, err := store.GetFile(ctx, childFile.ID); !errors.Is(err, usecasefile.ErrFileNotFound) {
		t.Errorf("childFile GetFile = %v, want ErrFileNotFound", err)
	}
}

func TestDeleteFolderCascadeMissing(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if _, err := store.DeleteFolderCascade(ctx, uuid.NewV7()); !errors.Is(err, usecasefolder.ErrFolderNotFound) {
		t.Errorf("DeleteFolderCascade = %v, want ErrFolderNotFound", err)
	}
}
