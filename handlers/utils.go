package handlers

import (
	"db-park/models"
	"encoding/json"
	htmux "github.com/dimfeld/httptreemux"
	"net/http"
	"path"
	"strings"
)

func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

func Get404(w http.ResponseWriter, what string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	page := models.NotFoundPage{what}
	_ = json.NewEncoder(w).Encode(page)
}

func getRootPage(writer http.ResponseWriter, request *http.Request, ps map[string]string) {
	writer.WriteHeader(200)
}

func RootHandler(router *htmux.TreeMux) {
	router.GET("/", getRootPage)
}