package web

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"uuid"

	usecasefile "alp-drive/internal/application/usecases/file"
	usecasefolder "alp-drive/internal/application/usecases/folder"
	"alp-drive/internal/domain/file"
	"alp-drive/internal/domain/folder"
)

type fakeFolderLister struct {
	folders []folder.Folder
	err     error
}

func (f fakeFolderLister) Execute(context.Context, usecasefolder.ListFoldersInput) ([]folder.Folder, error) {
	return f.folders, f.err
}

type fakeFolderCreator struct {
	created folder.Folder
	err     error
}

func (f fakeFolderCreator) Execute(_ context.Context, input usecasefolder.CreateFolderInput) (folder.Folder, error) {
	if err := input.Validate(); err != nil {
		return folder.Folder{}, err
	}
	return f.created, f.err
}

type fakeFolderRenamer struct{ err error }

func (f fakeFolderRenamer) Execute(context.Context, usecasefolder.RenameFolderInput) (folder.Folder, error) {
	return folder.Folder{}, f.err
}

type fakeFolderMover struct{ err error }

func (f fakeFolderMover) Execute(context.Context, usecasefolder.MoveFolderInput) (folder.Folder, error) {
	return folder.Folder{}, f.err
}

type fakeFolderDeleter struct{ err error }

func (f fakeFolderDeleter) Execute(context.Context, usecasefolder.DeleteFolderInput) error {
	return f.err
}

type fakeFileLister struct {
	files []file.File
	err   error
}

func (f fakeFileLister) Execute(context.Context, usecasefile.ListFilesByFolderInput) ([]file.File, error) {
	return f.files, f.err
}

type fakeFileCreator struct {
	created file.File
	err     error
}

func (f fakeFileCreator) Execute(context.Context, usecasefile.CreateFileInput) (file.File, error) {
	return f.created, f.err
}

type fakeFileRenamer struct{ err error }

func (f fakeFileRenamer) Execute(context.Context, usecasefile.RenameFileInput) (file.File, error) {
	return file.File{}, f.err
}

type fakeFileMover struct{ err error }

func (f fakeFileMover) Execute(context.Context, usecasefile.MoveFileInput) (file.File, error) {
	return file.File{}, f.err
}

type fakeFileDeleter struct{ err error }

func (f fakeFileDeleter) Execute(context.Context, usecasefile.DeleteFileInput) error {
	return f.err
}

type fakeFileDownloader struct {
	out usecasefile.DownloadFileOutput
	err error
}

func (f fakeFileDownloader) Execute(context.Context, usecasefile.DownloadFileInput) (usecasefile.DownloadFileOutput, error) {
	return f.out, f.err
}

type routerDeps struct {
	folders      fakeFolderLister
	createFolder fakeFolderCreator
	renameFolder fakeFolderRenamer
	moveFolder   fakeFolderMover
	deleteFolder fakeFolderDeleter
	files        fakeFileLister
	createFile   fakeFileCreator
	renameFile   fakeFileRenamer
	moveFile     fakeFileMover
	deleteFile   fakeFileDeleter
	downloadFile fakeFileDownloader
}

func newTestRouter(d routerDeps) http.Handler {
	h := NewDriveHandler(
		d.folders, d.createFolder, d.renameFolder, d.moveFolder, d.deleteFolder,
		d.files, d.createFile, d.renameFile, d.moveFile, d.deleteFile, d.downloadFile,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	return NewRouter(h)
}

var htmx = map[string]string{"HX-Request": "true"}

func do(t *testing.T, router http.Handler, method, path string, form url.Values, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req := httptest.NewRequest(method, path, body)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestList_RootPageOrFragment(t *testing.T) {
	router := newTestRouter(routerDeps{folders: fakeFolderLister{folders: []folder.Folder{{ID: uuid.NewV7(), Name: "Projects"}}}})

	page := do(t, router, "GET", "/folders", nil, nil)
	frag := do(t, router, "GET", "/folders", nil, htmx)

	if !strings.Contains(page.Body.String(), "<html") || !strings.Contains(page.Body.String(), "Projects") {
		t.Errorf("page: want full document with the folder row")
	}
	if strings.Contains(frag.Body.String(), "<html") || !strings.Contains(frag.Body.String(), "Projects") {
		t.Errorf("fragment: want the browser only, got %q", frag.Body.String()[:min(80, len(frag.Body.String()))])
	}
	if page.Header().Get("Vary") != "HX-Request" {
		t.Errorf("Vary = %q", page.Header().Get("Vary"))
	}
}

func TestList_FolderNotFound(t *testing.T) {
	router := newTestRouter(routerDeps{})
	rec := do(t, router, "GET", "/folders/not-a-uuid", nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestCreateFolder_ValidationError(t *testing.T) {
	router := newTestRouter(routerDeps{})
	rec := do(t, router, "POST", "/folders/new", url.Values{"name": {"  "}}, htmx)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Enter a name.") {
		t.Errorf("missing field error, got %q", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "<html") {
		t.Errorf("htmx validation error: want the dialog fragment only")
	}
}

func TestCreateFolder_SuccessRedirect(t *testing.T) {
	router := newTestRouter(routerDeps{createFolder: fakeFolderCreator{created: folder.Folder{ID: uuid.NewV7(), Name: "New"}}})

	h := do(t, router, "POST", "/folders/new", url.Values{"name": {"New"}}, htmx)
	if h.Code != http.StatusNoContent || h.Header().Get("HX-Redirect") != "/folders" {
		t.Errorf("htmx: %d %q", h.Code, h.Header().Get("HX-Redirect"))
	}

	b := do(t, router, "POST", "/folders/new", url.Values{"name": {"New"}}, nil)
	if b.Code != http.StatusSeeOther || b.Header().Get("Location") != "/folders" {
		t.Errorf("browser: %d %q", b.Code, b.Header().Get("Location"))
	}
}

func TestCreateFolder_UnknownErrorIsRetargeted(t *testing.T) {
	router := newTestRouter(routerDeps{createFolder: fakeFolderCreator{err: errors.New("disk full")}})

	rec := do(t, router, "POST", "/folders/new", url.Values{"name": {"New"}}, htmx)
	if rec.Code != 500 || rec.Header().Get("HX-Retarget") != "#alerts" || rec.Header().Get("HX-Reswap") != "innerHTML" {
		t.Errorf("%d %q %q", rec.Code, rec.Header().Get("HX-Retarget"), rec.Header().Get("HX-Reswap"))
	}
	if !strings.Contains(rec.Body.String(), `role="alert"`) || strings.Contains(rec.Body.String(), "disk full") {
		t.Errorf("want a generic alert, never the detail: %q", rec.Body.String())
	}
}

func TestDeleteFolder_Redirects(t *testing.T) {
	router := newTestRouter(routerDeps{})
	id := uuid.NewV7().String()

	rec := do(t, router, "POST", "/folders/"+id+"/delete", url.Values{}, nil)
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/folders" {
		t.Errorf("%d %q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestDownloadFile_ServesContent(t *testing.T) {
	f := file.File{ID: uuid.NewV7(), Name: "report.pdf", MimeType: "application/pdf", Size: 5, UpdatedAt: time.Now()}
	router := newTestRouter(routerDeps{
		downloadFile: fakeFileDownloader{out: usecasefile.DownloadFileOutput{File: f, Content: io.NopCloser(strings.NewReader("hello"))}},
	})

	rec := do(t, router, "GET", "/files/"+f.ID.String()+"/download", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/pdf" {
		t.Errorf("Content-Type = %q", rec.Header().Get("Content-Type"))
	}
	if !strings.Contains(rec.Header().Get("Content-Disposition"), "report.pdf") {
		t.Errorf("Content-Disposition = %q", rec.Header().Get("Content-Disposition"))
	}
	if rec.Body.String() != "hello" {
		t.Errorf("body = %q", rec.Body.String())
	}
}

func TestCrossOriginPostIsRejected(t *testing.T) {
	router := newTestRouter(routerDeps{createFolder: fakeFolderCreator{created: folder.Folder{ID: uuid.NewV7()}}})

	rec := do(t, router, "POST", "/folders/new", url.Values{"name": {"New"}}, map[string]string{"Sec-Fetch-Site": "cross-site"})
	same := do(t, router, "POST", "/folders/new", url.Values{"name": {"New"}}, map[string]string{"Sec-Fetch-Site": "same-origin"})

	if rec.Code != http.StatusForbidden {
		t.Errorf("cross-site = %d, want 403", rec.Code)
	}
	if same.Code != http.StatusSeeOther {
		t.Errorf("same-origin = %d, want 303", same.Code)
	}
}
