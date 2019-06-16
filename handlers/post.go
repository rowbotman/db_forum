package handlers

import (
	"fmt"
	json "github.com/mailru/easyjson"
	"github.com/naoina/denco"
	"../db"
	"../models"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func PostChangeInfo(w http.ResponseWriter, req *http.Request, ps denco.Params) {
	log.Println("post change info:", req.RequestURI)
	var data models.DataForUpdPost
	var err error
	_ = json.UnmarshalFromReader(req.Body, &data)
	id := int64(0)
	if postId := ps.Get("id"); len(postId) <= 0 {
		http.Error(w, "Can't parse id", http.StatusBadRequest)
		return
	} else {
		id, err = strconv.ParseInt(postId, 10, 64)
		if err != nil {
			http.Error(w, "Can't parse id", http.StatusBadRequest)
			return
		}
	}
	data.Id = id
	forum, err := db.UpdatePost(data)
	if err != nil {
		if forum.Uid == -1  {
			Get404(w, err.Error())
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	output, err := json.Marshal(forum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(output)
}

func PostGetInfo(w http.ResponseWriter, req *http.Request, ps denco.Params) {
	log.Println("post get info:", req.RequestURI)
	id := int64(0)
	var err error
	if postId := ps.Get("id"); len(postId) <= 0 {
		http.Error(w, "Can't parse id", http.StatusBadRequest)
		return
	} else {
		id, err = strconv.ParseInt(postId, 10, 64)
		if err != nil {
			http.Error(w, "Can't parse id", http.StatusBadRequest)
			return
		}
	}

	_ = req.ParseForm() // parses request body and query and stores result in r.Form
	var array []string
	array = strings.Split(req.FormValue("related"), ",")
	details, err := db.GetPostInfo(id, array, w)
	if err != nil {
		if details["err"] == true {
			Get404(w, err.Error())
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func PostHandler(router **denco.Mux) []denco.Handler {
	fmt.Println("posts handlers initialized")
	return []denco.Handler{
		(*router).POST("/api/post/:id/details", PostChangeInfo),
		(*router).GET( "/api/post/:id/details", PostGetInfo)}
}
