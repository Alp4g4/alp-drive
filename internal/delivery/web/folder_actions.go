package web

import (
	"errors"
	"net/http"
	"uuid"

	usecasefolder "alp-drive/internal/application/usecases/folder"
	"alp-drive/internal/delivery/web/views"
	domainfolder "alp-drive/internal/domain/folder"
)

func (h *DriveHandler) NewFolderDialog(w http.ResponseWriter, r *http.Request) {
	in := r.URL.Query().Get("in")
	crumbs := queryCrumbs(r, "c")
	form := views.NewFolderForm{PostURL: newFolderURL(in, crumbs), CancelURL: folderListURL(in, crumbs)}
	h.renderWithDialog(w, r, http.StatusOK, in, crumbs, views.NewFolderDialog(form))
}

func (h *DriveHandler) CreateFolder(w http.ResponseWriter, r *http.Request) {
	in := r.URL.Query().Get("in")
	crumbs := queryCrumbs(r, "c")
	form := views.NewFolderForm{
		Name:      r.PostFormValue("name"),
		PostURL:   newFolderURL(in, crumbs),
		CancelURL: folderListURL(in, crumbs),
	}

	var parentID *uuid.UUID
	if in != "" {
		id, err := uuid.Parse(in)
		if err != nil {
			renderError(w, r, http.StatusNotFound, "This folder does not exist.")
			return
		}
		parentID = &id
	}

	_, err := h.createFolder.Execute(r.Context(), usecasefolder.CreateFolderInput{Name: form.Name, ParentID: parentID})
	if err != nil {
		if presentFolderFormError(&form.NameError, err) {
			h.renderWithDialog(w, r, http.StatusUnprocessableEntity, in, crumbs, views.NewFolderDialog(form))
			return
		}
		h.logger.Error("creating folder", "error", err)
		renderError(w, r, http.StatusInternalServerError, "The folder could not be created.")
		return
	}

	redirect(w, r, folderListURL(in, crumbs))
}

func (h *DriveHandler) RenameFolderDialog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	in := r.URL.Query().Get("in")
	crumbs := queryCrumbs(r, "c")
	name := r.URL.Query().Get("name")
	form := views.RenameForm{Title: "Rename folder", Name: name, PostURL: renameFolderURL(id, in, crumbs, name), CancelURL: folderListURL(in, crumbs)}
	h.renderWithDialog(w, r, http.StatusOK, in, crumbs, views.RenameDialog(form))
}

func (h *DriveHandler) RenameFolder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	in := r.URL.Query().Get("in")
	crumbs := queryCrumbs(r, "c")
	originalName := r.URL.Query().Get("name")
	form := views.RenameForm{
		Title:     "Rename folder",
		Name:      r.PostFormValue("name"),
		PostURL:   renameFolderURL(id, in, crumbs, originalName),
		CancelURL: folderListURL(in, crumbs),
	}

	folderID, err := uuid.Parse(id)
	if err != nil {
		renderError(w, r, http.StatusNotFound, "This folder does not exist.")
		return
	}

	_, err = h.renameFolder.Execute(r.Context(), usecasefolder.RenameFolderInput{FolderID: folderID, Name: form.Name})
	if err != nil {
		if presentFolderFormError(&form.NameError, err) {
			h.renderWithDialog(w, r, http.StatusUnprocessableEntity, in, crumbs, views.RenameDialog(form))
			return
		}
		h.logger.Error("renaming folder", "error", err)
		renderError(w, r, http.StatusInternalServerError, "The folder could not be renamed.")
		return
	}

	redirect(w, r, folderListURL(in, crumbs))
}

func (h *DriveHandler) DeleteFolderDialog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	in := r.URL.Query().Get("in")
	crumbs := queryCrumbs(r, "c")
	form := views.DeleteConfirmForm{
		Title:     "Delete folder?",
		Message:   "This deletes the folder and everything inside it. This can't be undone.",
		PostURL:   deleteFolderURL(id, in, crumbs),
		CancelURL: folderListURL(in, crumbs),
	}
	h.renderWithDialog(w, r, http.StatusOK, in, crumbs, views.DeleteConfirmDialog(form))
}

func (h *DriveHandler) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	in := r.URL.Query().Get("in")
	crumbs := queryCrumbs(r, "c")

	folderID, err := uuid.Parse(id)
	if err != nil {
		renderError(w, r, http.StatusNotFound, "This folder does not exist.")
		return
	}

	if err := h.deleteFolder.Execute(r.Context(), usecasefolder.DeleteFolderInput{FolderID: folderID}); err != nil {
		h.logger.Error("deleting folder", "error", err)
		renderError(w, r, http.StatusInternalServerError, "The folder could not be deleted.")
		return
	}

	redirect(w, r, folderListURL(in, crumbs))
}

func (h *DriveHandler) MoveFolderDialog(w http.ResponseWriter, r *http.Request) {
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
		if f.ID.String() == id {
			continue // a folder cannot be moved into itself
		}
		next := withCrumb(targetCrumbs, f.ID.String(), f.Name)
		targets = append(targets, views.MoveTargetFolder{Name: f.Name, BrowseURL: moveFolderURL(id, in, crumbs, f.ID.String(), next)})
	}

	breadcrumb := buildMoveBreadcrumb(targetCrumbs, func(tid string, tc []crumb) string {
		return moveFolderURL(id, in, crumbs, tid, tc)
	})

	form := views.MoveForm{
		Title:      "Move folder",
		Breadcrumb: breadcrumb,
		Targets:    targets,
		Empty:      len(targets) == 0,
		PostURL:    moveFolderURL(id, in, crumbs, target, targetCrumbs),
		CancelURL:  folderListURL(in, crumbs),
	}
	h.renderWithDialog(w, r, http.StatusOK, in, crumbs, views.MoveDialog(form))
}

func (h *DriveHandler) MoveFolder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	in := r.URL.Query().Get("in")
	crumbs := queryCrumbs(r, "c")
	target := r.URL.Query().Get("target")

	folderID, err := uuid.Parse(id)
	if err != nil {
		renderError(w, r, http.StatusNotFound, "This folder does not exist.")
		return
	}

	var newParentID *uuid.UUID
	if target != "" {
		tid, err := uuid.Parse(target)
		if err != nil {
			renderError(w, r, http.StatusNotFound, "This destination does not exist.")
			return
		}
		newParentID = &tid
	}

	if _, err := h.moveFolder.Execute(r.Context(), usecasefolder.MoveFolderInput{FolderID: folderID, NewParentID: newParentID}); err != nil {
		h.logger.Error("moving folder", "error", err)
		renderError(w, r, http.StatusInternalServerError, "The folder could not be moved there.")
		return
	}

	redirect(w, r, folderListURL(in, crumbs))
}

func presentFolderFormError(nameError *string, err error) bool {
	switch {
	case errors.Is(err, usecasefolder.ErrFolderNameEmpty):
		*nameError = "Enter a name."
	case errors.Is(err, domainfolder.ErrFolderCannotBeOwnParent):
		*nameError = "A folder can't contain itself."
	default:
		return false
	}
	return true
}

// buildMoveBreadcrumb builds the breadcrumb of the move dialog's own,
// independent browsing state (never the page's breadcrumb, which stays in
// place behind it). urlFor builds the link for browsing to that ancestor.
func buildMoveBreadcrumb(targetCrumbs []crumb, urlFor func(targetID string, tc []crumb) string) []views.BreadcrumbItem {
	items := make([]views.BreadcrumbItem, 0, len(targetCrumbs)+1)
	items = append(items, views.BreadcrumbItem{Name: "Home", URL: urlFor("", nil), Current: len(targetCrumbs) == 0})
	for i, c := range targetCrumbs {
		items = append(items, views.BreadcrumbItem{
			Name:    c.Name,
			URL:     urlFor(c.ID, targetCrumbs[:i+1]),
			Current: i == len(targetCrumbs)-1,
		})
	}
	return items
}
