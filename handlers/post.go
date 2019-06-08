package handlers

import (
	"../db"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)

func postChangeInfo(w http.ResponseWriter, req *http.Request) {
	var data db.DataForUpdPost
	_= json.NewDecoder(req.Body).Decode(&data)
	params := mux.Vars(req)
	id := int64(0)
	if postId, ok := params["id"]; !ok {
		http.Error(w, "Can't parse id", http.StatusBadRequest)
	} else {
		id, _ = strconv.ParseInt(postId, 10, 64)
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

func PostGetInfo(w http.ResponseWriter,req *http.Request) {
	params := mux.Vars(req)
	id := int64(0)
	var err error
	if postId, ok := params["id"]; !ok {
		http.Error(w, "Can't parse id", http.StatusBadRequest)
		return
	} else {
		id, err = strconv.ParseInt(postId, 10, 64)
		if err !=  nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	_ = req.ParseForm() // parses request body and query and stores result in r.Form
	var array []string
	array = strings.Split(req.FormValue("related"), ",")
	details, err := db.GetPostInfo(id, array)
	if err != nil {
		if details["err"] == true {
			Get404(w, err.Error())
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output, err := json.Marshal(details)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(output)
}

func PostHandler(router **mux.Router) {
	fmt.Println("posts handlers initialized")
	(*router).HandleFunc("/api/post/{id}/details", postChangeInfo).Methods("POST")
	(*router).HandleFunc("/api/post/{id}/details", PostGetInfo).Methods("GET")
}
