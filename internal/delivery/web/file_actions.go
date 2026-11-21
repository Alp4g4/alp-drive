package web

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"uuid"

	usecasefile "alp-drive/internal/application/usecases/file"
	usecasefolder "alp-drive/internal/application/usecases/folder"
	"alp-drive/internal/delivery/web/views"
)

const maxUploadSize = 1 << 30 // 1 GiB

func (h *DriveHandler) UploadDialog(w http.ResponseWriter, r *http.Request) {
	in := r.URL.Query().Get("in")
	crumbs := queryCrumbs(r, "c")
	form := views.UploadForm{PostURL: uploadURL(in, crumbs), CancelURL: folderListURL(in, crumbs)}
	h.renderWithDialog(w, r, http.StatusOK, in, crumbs, views.UploadDialog(form))
}

func (h *DriveHandler) CreateFile(w http.ResponseWriter, r *http.Request) {
	in := r.URL.Query().Get("in")
	crumbs := queryCrumbs(r, "c")
	form := views.UploadForm{PostURL: uploadURL(in, crumbs), CancelURL: folderListURL(in, crumbs)}

	folderID, err := uuid.Parse(in)
	if err != nil {
		renderError(w, r, http.StatusNotFound, "This folder does not exist.")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	file, header, err := r.FormFile("file")
	if err != nil {
		form.FileError = "Choose a file."
		h.renderWithDialog(w, r, http.StatusUnprocessableEntity, in, crumbs, views.UploadDialog(form))
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(header.Filename))
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	_, err = h.createFile.Execute(r.Context(), usecasefile.CreateFileInput{
		Name:     header.Filename,
		FolderID: folderID,
		Content:  file,
		Size:     header.Size,
		MimeType: mimeType,
	})
	if err != nil {
		if presentFileFormError(&form.FileError, err) {
			h.renderWithDialog(w, r, http.StatusUnprocessableEntity, in, crumbs, views.UploadDialog(form))
			return
		}
		h.logger.Error("uploading file", "error", err)
		renderError(w, r, http.StatusInternalServerError, "The file could not be uploaded.")
		return
	}

	redirect(w, r, folderListURL(in, crumbs))
}

func (h *DriveHandler) RenameFileDialog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	in := r.URL.Query().Get("in")
	crumbs := queryCrumbs(r, "c")
	name := r.URL.Query().Get("name")
	form := views.RenameForm{Title: "Rename file", Name: name, PostURL: renameFileURL(id, in, crumbs, name), CancelURL: folderListURL(in, crumbs)}
	h.renderWithDialog(w, r, http.StatusOK, in, crumbs, views.RenameDialog(form))
}

func (h *DriveHandler) RenameFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	in := r.URL.Query().Get("in")
	crumbs := queryCrumbs(r, "c")
	originalName := r.URL.Query().Get("name")
	form := views.RenameForm{
		Title:     "Rename file",
		Name:      r.PostFormValue("name"),
		PostURL:   renameFileURL(id, in, crumbs, originalName),
		CancelURL: folderListURL(in, crumbs),
	}

	fileID, err := uuid.Parse(id)
	if err != nil {
		renderError(w, r, http.StatusNotFound, "This file does not exist.")
		return
	}

	_, err = h.renameFile.Execute(r.Context(), usecasefile.RenameFileInput{FileID: fileID, Name: form.Name})
	if err != nil {
		if presentFileFormError(&form.NameError, err) {
			h.renderWithDialog(w, r, http.StatusUnprocessableEntity, in, crumbs, views.RenameDialog(form))
			return
		}
		h.logger.Error("renaming file", "error", err)
		renderError(w, r, http.StatusInternalServerError, "The file could not be renamed.")
		return
	}

	redirect(w, r, folderListURL(in, crumbs))
}

func (h *DriveHandler) DeleteFileDialog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	in := r.URL.Query().Get("in")
	crumbs := queryCrumbs(r, "c")
	form := views.DeleteConfirmForm{
		Title:     "Delete file?",
		Message:   "This can't be undone.",
		PostURL:   deleteFileURL(id, in, crumbs),
		CancelURL: folderListURL(in, crumbs),
	}
	h.renderWithDialog(w, r, http.StatusOK, in, crumbs, views.DeleteConfirmDialog(form))
}

func (h *DriveHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	in := r.URL.Query().Get("in")
	crumbs := queryCrumbs(r, "c")

	fileID, err := uuid.Parse(id)
	if err != nil {
		renderError(w, r, http.StatusNotFound, "This file does not exist.")
		return
	}

	if err := h.deleteFile.Execute(r.Context(), usecasefile.DeleteFileInput{FileID: fileID}); err != nil {
		h.logger.Error("deleting file", "error", err)
		renderError(w, r, http.StatusInternalServerError, "The file could not be deleted.")
		return
	}

	redirect(w, r, folderListURL(in, crumbs))
}

func (h *DriveHandler) MoveFileDialog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	in := r.URL.Query().Get("in")
	crumbs := queryCrumbs(r, "c")
	target := r.URL.Query().Get("target")
	targetCrumbs := queryCrumbs(r, "tc")

	var targetID *uuid.UUID
	if target != "" {
		tid, err := uuid.Parse(target)
		if err != nil {
			renderError(w, r, http.StatusNotFound, "This destination does not exist.")
			return
		}
		targetID = &tid
	}

	folders, err := h.listFolders.Execute(r.Context(), usecasefolder.ListFoldersInput{ParentID: targetID})
	if err != nil {
		h.logger.Error("listing move destinations", "error", err)
		renderError(w, r, http.StatusInternalServerError, "The destinations could not be loaded.")
		return
	}

	targets := make([]views.MoveTargetFolder, 0, len(folders))
	for _, f := range folders {
		next := withCrumb(targetCrumbs, f.ID.String(), f.Name)
		targets = append(targets, views.MoveTargetFolder{Name: f.Name, BrowseURL: moveFileURL(id, in, crumbs, f.ID.String(), next)})
	}

	breadcrumb := buildMoveBreadcrumb(targetCrumbs, func(tid string, tc []crumb) string {
		return moveFileURL(id, in, crumbs, tid, tc)
	})

	form := views.MoveForm{
		Title:      "Move file",
		Breadcrumb: breadcrumb,
		Targets:    targets,
		Empty:      len(targets) == 0,
		PostURL:    moveFileURL(id, in, crumbs, target, targetCrumbs),
		CancelURL:  folderListURL(in, crumbs),
	}
	h.renderWithDialog(w, r, http.StatusOK, in, crumbs, views.MoveDialog(form))
}

func (h *DriveHandler) MoveFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	in := r.URL.Query().Get("in")
	crumbs := queryCrumbs(r, "c")
	target := r.URL.Query().Get("target")

	fileID, err := uuid.Parse(id)
	if err != nil {
		renderError(w, r, http.StatusNotFound, "This file does not exist.")
		return
	}

	newFolderID, err := uuid.Parse(target)
	if err != nil {
		renderError(w, r, http.StatusUnprocessableEntity, "Choose a destination folder.")
		return
	}

	if _, err := h.moveFile.Execute(r.Context(), usecasefile.MoveFileInput{FileID: fileID, NewFolderID: newFolderID}); err != nil {
		h.logger.Error("moving file", "error", err)
		renderError(w, r, http.StatusInternalServerError, "The file could not be moved there.")
		return
	}

	redirect(w, r, folderListURL(in, crumbs))
}

func (h *DriveHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	fileID, err := uuid.Parse(id)
	if err != nil {
		renderError(w, r, http.StatusNotFound, "This file does not exist.")
		return
	}

	out, err := h.download.Execute(r.Context(), usecasefile.DownloadFileInput{FileID: fileID})
	if err != nil {
		h.logger.Error("downloading file", "error", err)
		renderError(w, r, http.StatusInternalServerError, "The file could not be downloaded.")
		return
	}
	defer out.Content.Close()

	w.Header().Set("Content-Disposition", `attachment; filename="`+strings.ReplaceAll(out.File.Name, `"`, "")+`"`)
	w.Header().Set("Content-Type", out.File.MimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(out.File.Size, 10))
	_, _ = io.Copy(w, out.Content)
}

func presentFileFormError(fieldError *string, err error) bool {
	switch {
	case errors.Is(err, usecasefile.ErrFileNameEmpty):
		*fieldError = "Enter a name."
	default:
		return false
	}
	return true
}
