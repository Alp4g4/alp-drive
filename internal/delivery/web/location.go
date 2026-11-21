package web

import (
	"net/http"
	"net/url"
	"strings"
)

type crumb struct {
	ID   string
	Name string
}

func parseCrumbs(values []string) []crumb {
	crumbs := make([]crumb, 0, len(values))
	for _, v := range values {
		id, name, ok := strings.Cut(v, "|")
		if !ok || id == "" {
			continue
		}
		crumbs = append(crumbs, crumb{ID: id, Name: name})
	}
	return crumbs
}

func encodeCrumbs(q url.Values, key string, crumbs []crumb) {
	for _, c := range crumbs {
		q.Add(key, c.ID+"|"+c.Name)
	}
}

func withCrumb(crumbs []crumb, id, name string) []crumb {
	next := make([]crumb, len(crumbs), len(crumbs)+1)
	copy(next, crumbs)
	return append(next, crumb{ID: id, Name: name})
}

func folderPath(id string) string {
	if id == "" {
		return "/folders"
	}
	return "/folders/" + id
}

func folderListURL(id string, crumbs []crumb) string {
	q := url.Values{}
	encodeCrumbs(q, "c", crumbs)
	return withQuery(folderPath(id), q)
}

func newFolderURL(in string, crumbs []crumb) string {
	q := returnQuery(in, crumbs)
	return withQuery("/folders/new", q)
}

func renameFolderURL(id, in string, crumbs []crumb) string {
	q := returnQuery(in, crumbs)
	return withQuery("/folders/"+id+"/rename", q)
}

func deleteFolderURL(id, in string, crumbs []crumb) string {
	q := returnQuery(in, crumbs)
	return withQuery("/folders/"+id+"/delete", q)
}

func moveFolderURL(id, in string, crumbs []crumb, target string, targetCrumbs []crumb) string {
	q := returnQuery(in, crumbs)
	q.Set("target", target)
	encodeCrumbs(q, "tc", targetCrumbs)
	return withQuery("/folders/"+id+"/move", q)
}

func uploadURL(in string, crumbs []crumb) string {
	q := returnQuery(in, crumbs)
	return withQuery("/files/new", q)
}

func renameFileURL(id, in string, crumbs []crumb) string {
	q := returnQuery(in, crumbs)
	return withQuery("/files/"+id+"/rename", q)
}

func deleteFileURL(id, in string, crumbs []crumb) string {
	q := returnQuery(in, crumbs)
	return withQuery("/files/"+id+"/delete", q)
}

func moveFileURL(id, in string, crumbs []crumb, target string, targetCrumbs []crumb) string {
	q := returnQuery(in, crumbs)
	q.Set("target", target)
	encodeCrumbs(q, "tc", targetCrumbs)
	return withQuery("/files/"+id+"/move", q)
}

func downloadFileURL(id string) string {
	return "/files/" + id + "/download"
}

func returnQuery(in string, crumbs []crumb) url.Values {
	q := url.Values{}
	if in != "" {
		q.Set("in", in)
	}
	encodeCrumbs(q, "c", crumbs)
	return q
}

func withQuery(path string, q url.Values) string {
	if len(q) == 0 {
		return path
	}
	return path + "?" + q.Encode()
}

func queryCrumbs(r *http.Request, key string) []crumb {
	return parseCrumbs(r.URL.Query()[key])
}
