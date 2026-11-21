package web

import (
	"net/http"

	"alp-drive/internal/delivery/web/static"
)

func NewRouter(drive *DriveHandler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", static.Handler())

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/folders", http.StatusSeeOther)
	})

	mux.HandleFunc("GET /folders", drive.List)
	mux.HandleFunc("GET /folders/{id}", drive.List)
	mux.HandleFunc("GET /folders/new", drive.NewFolderDialog)
	mux.HandleFunc("POST /folders/new", drive.CreateFolder)
	mux.HandleFunc("GET /folders/{id}/rename", drive.RenameFolderDialog)
	mux.HandleFunc("POST /folders/{id}/rename", drive.RenameFolder)
	mux.HandleFunc("GET /folders/{id}/delete", drive.DeleteFolderDialog)
	mux.HandleFunc("POST /folders/{id}/delete", drive.DeleteFolder)
	mux.HandleFunc("GET /folders/{id}/move", drive.MoveFolderDialog)
	mux.HandleFunc("POST /folders/{id}/move", drive.MoveFolder)

	mux.HandleFunc("GET /files/new", drive.UploadDialog)
	mux.HandleFunc("POST /files/new", drive.CreateFile)
	mux.HandleFunc("GET /files/{id}/rename", drive.RenameFileDialog)
	mux.HandleFunc("POST /files/{id}/rename", drive.RenameFile)
	mux.HandleFunc("GET /files/{id}/delete", drive.DeleteFileDialog)
	mux.HandleFunc("POST /files/{id}/delete", drive.DeleteFile)
	mux.HandleFunc("GET /files/{id}/move", drive.MoveFileDialog)
	mux.HandleFunc("POST /files/{id}/move", drive.MoveFile)
	mux.HandleFunc("GET /files/{id}/download", drive.DownloadFile)

	return http.NewCrossOriginProtection().Handler(mux)
}
