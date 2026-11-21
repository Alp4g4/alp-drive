package web

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"uuid"

	"github.com/a-h/templ"

	usecasefile "alp-drive/internal/application/usecases/file"
	usecasefolder "alp-drive/internal/application/usecases/folder"
	"alp-drive/internal/delivery/web/views"
	"alp-drive/internal/delivery/web/views/components"
	"alp-drive/internal/domain/file"
	"alp-drive/internal/domain/folder"
)

type folderLister interface {
	Execute(ctx context.Context, input usecasefolder.ListFoldersInput) ([]folder.Folder, error)
}

type folderCreator interface {
	Execute(ctx context.Context, input usecasefolder.CreateFolderInput) (folder.Folder, error)
}

type folderRenamer interface {
	Execute(ctx context.Context, input usecasefolder.RenameFolderInput) (folder.Folder, error)
}

type folderMover interface {
	Execute(ctx context.Context, input usecasefolder.MoveFolderInput) (folder.Folder, error)
}

type folderDeleter interface {
	Execute(ctx context.Context, input usecasefolder.DeleteFolderInput) error
}

type fileLister interface {
	Execute(ctx context.Context, input usecasefile.ListFilesByFolderInput) ([]file.File, error)
}

type fileCreator interface {
	Execute(ctx context.Context, input usecasefile.CreateFileInput) (file.File, error)
}

type fileRenamer interface {
	Execute(ctx context.Context, input usecasefile.RenameFileInput) (file.File, error)
}

type fileMover interface {
	Execute(ctx context.Context, input usecasefile.MoveFileInput) (file.File, error)
}

type fileDeleter interface {
	Execute(ctx context.Context, input usecasefile.DeleteFileInput) error
}

type fileDownloader interface {
	Execute(ctx context.Context, input usecasefile.DownloadFileInput) (usecasefile.DownloadFileOutput, error)
}

type DriveHandler struct {
	listFolders  folderLister
	createFolder folderCreator
	renameFolder folderRenamer
	moveFolder   folderMover
	deleteFolder folderDeleter

	listFiles  fileLister
	createFile fileCreator
	renameFile fileRenamer
	moveFile   fileMover
	deleteFile fileDeleter
	download   fileDownloader

	logger *slog.Logger
}

func NewDriveHandler(
	listFolders folderLister,
	createFolder folderCreator,
	renameFolder folderRenamer,
	moveFolder folderMover,
	deleteFolder folderDeleter,
	listFiles fileLister,
	createFile fileCreator,
	renameFile fileRenamer,
	moveFile fileMover,
	deleteFile fileDeleter,
	download fileDownloader,
	logger *slog.Logger,
) *DriveHandler {
	return &DriveHandler{
		listFolders:  listFolders,
		createFolder: createFolder,
		renameFolder: renameFolder,
		moveFolder:   moveFolder,
		deleteFolder: deleteFolder,
		listFiles:    listFiles,
		createFile:   createFile,
		renameFile:   renameFile,
		moveFile:     moveFile,
		deleteFile:   deleteFile,
		download:     download,
		logger:       logger,
	}
}

// List renders the contents of a folder: GET /folders (root) or GET /folders/{id}.
func (h *DriveHandler) List(w http.ResponseWriter, r *http.Request) {
	currentID := r.PathValue("id")
	crumbs := queryCrumbs(r, "c")

	view, ok := h.loadDriveView(w, r, currentID, crumbs)
	if !ok {
		return
	}

	render(w, r, http.StatusOK, views.DrivePage(driveTitle(crumbs), view, nil), views.DriveBrowser(view))
}

// loadDriveView loads the folders and files of currentID and presents them.
// It renders the error itself and reports false when it could not load them,
// so the caller only has to return.
func (h *DriveHandler) loadDriveView(w http.ResponseWriter, r *http.Request, currentID string, crumbs []crumb) (views.DriveView, bool) {
	var parentID *uuid.UUID
	if currentID != "" {
		id, err := uuid.Parse(currentID)
		if err != nil {
			renderError(w, r, http.StatusNotFound, "This folder does not exist.")
			return views.DriveView{}, false
		}
		parentID = &id
	}

	folders, err := h.listFolders.Execute(r.Context(), usecasefolder.ListFoldersInput{ParentID: parentID})
	if err != nil {
		h.logger.Error("listing folders", "error", err)
		renderError(w, r, http.StatusInternalServerError, "The folders could not be loaded.")
		return views.DriveView{}, false
	}

	var files []file.File
	if parentID != nil {
		files, err = h.listFiles.Execute(r.Context(), usecasefile.ListFilesByFolderInput{FolderID: *parentID})
		if err != nil {
			h.logger.Error("listing files", "error", err)
			renderError(w, r, http.StatusInternalServerError, "The files could not be loaded.")
			return views.DriveView{}, false
		}
	}

	return h.presentDrive(currentID, crumbs, folders, files), true
}

// renderWithDialog renders the listing behind an open dialog: the full page
// for a browser, the dialog alone for htmx. The dialog's own form posts back
// to the same URL and is re-rendered this same way on a validation error.
func (h *DriveHandler) renderWithDialog(w http.ResponseWriter, r *http.Request, status int, currentID string, crumbs []crumb, dialog templ.Component) {
	view, ok := h.loadDriveView(w, r, currentID, crumbs)
	if !ok {
		return
	}

	render(w, r, status, views.DrivePage(driveTitle(crumbs), view, dialog), dialog)
}

func driveTitle(crumbs []crumb) string {
	if len(crumbs) == 0 {
		return "Home"
	}
	return crumbs[len(crumbs)-1].Name
}

func (h *DriveHandler) presentDrive(currentID string, crumbs []crumb, folders []folder.Folder, files []file.File) views.DriveView {
	breadcrumb := make([]views.BreadcrumbItem, 0, len(crumbs)+1)
	breadcrumb = append(breadcrumb, views.BreadcrumbItem{Name: "Home", URL: folderListURL("", nil), Current: len(crumbs) == 0})
	for i, c := range crumbs {
		breadcrumb = append(breadcrumb, views.BreadcrumbItem{
			Name:    c.Name,
			URL:     folderListURL(c.ID, crumbs[:i+1]),
			Current: i == len(crumbs)-1,
		})
	}

	folderRows := make([]views.FolderRow, len(folders))
	for i, f := range folders {
		id := f.ID.String()
		next := withCrumb(crumbs, id, f.Name)
		folderRows[i] = views.FolderRow{
			Name:      f.Name,
			OpenURL:   folderListURL(id, next),
			RenameURL: renameFolderURL(id, currentID, crumbs),
			MoveURL:   moveFolderURL(id, currentID, crumbs, "", nil),
			DeleteURL: deleteFolderURL(id, currentID, crumbs),
		}
	}

	fileRows := make([]views.FileRow, len(files))
	for i, f := range files {
		id := f.ID.String()
		fileRows[i] = views.FileRow{
			Name:            f.Name,
			SizeDisplay:     formatSize(f.Size),
			ModifiedDisplay: formatModified(f.UpdatedAt),
			Icon:            iconForMimeType(f.MimeType),
			DownloadURL:     downloadFileURL(id),
			RenameURL:       renameFileURL(id, currentID, crumbs),
			MoveURL:         moveFileURL(id, currentID, crumbs, "", nil),
			DeleteURL:       deleteFileURL(id, currentID, crumbs),
		}
	}

	return views.DriveView{
		Breadcrumb:   breadcrumb,
		CanUpload:    currentID != "",
		NewFolderURL: newFolderURL(currentID, crumbs),
		UploadURL:    uploadURL(currentID, crumbs),
		Folders:      folderRows,
		Files:        fileRows,
		Empty:        len(folders) == 0 && len(files) == 0,
	}
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func formatModified(t time.Time) string {
	return t.Format("2 Jan 2006, 15:04")
}

func iconForMimeType(mimeType string) components.IconKind {
	switch {
	case mimeType == "application/pdf":
		return components.IconPDF
	case strings.HasPrefix(mimeType, "image/"):
		return components.IconImage
	case strings.HasPrefix(mimeType, "text/"):
		return components.IconText
	case strings.Contains(mimeType, "spreadsheet") || strings.Contains(mimeType, "excel") || mimeType == "text/csv":
		return components.IconSheet
	default:
		return components.IconFile
	}
}
