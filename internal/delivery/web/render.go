package web

import (
	"bytes"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"

	"alp-drive/internal/delivery/web/views"
	"alp-drive/internal/delivery/web/views/components"
)

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-History-Restore-Request") != "true"
}

func render(w http.ResponseWriter, r *http.Request, status int, page, fragment templ.Component) {
	c := page
	if isHTMX(r) {
		c = fragment
	}

	var buf bytes.Buffer
	if err := c.Render(r.Context(), &buf); err != nil {
		slog.Error("rendering", "path", r.URL.Path, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Vary", "HX-Request")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

func redirect(w http.ResponseWriter, r *http.Request, url string) {
	if isHTMX(r) {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func renderError(w http.ResponseWriter, r *http.Request, status int, message string) {
	if isHTMX(r) {
		w.Header().Set("HX-Retarget", "#alerts")
		w.Header().Set("HX-Reswap", "innerHTML")
	}
	render(w, r, status, views.ErrorPage(message), components.Alert(components.AlertDanger, message))
}
