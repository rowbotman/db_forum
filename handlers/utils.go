package handlers

import (
	"encoding/json"
	"github.com/rowbotman/db_forum/models"
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