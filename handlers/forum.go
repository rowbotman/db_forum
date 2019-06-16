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
)

func forumCreate(w http.ResponseWriter, req *http.Request, _ denco.Params) {
	log.Println("forum create", req.RequestURI)
	var data models.DataForNewForum
	_ = json.UnmarshalFromReader(req.Body, &data)
	forum, err := db.InsertIntoForum(data)
	if err != nil {
		if len(forum.Slug) > 0 {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_, _, _ = json.MarshalToHTTPResponseWriter(forum, w)
			return
		}
		Get404(w, err.Error())
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _, _ = json.MarshalToHTTPResponseWriter(forum, w)
}

func forumGetInfo(w http.ResponseWriter,req *http.Request, ps denco.Params) {
	log.Println("forum get info", req.RequestURI)
	forumSlug := ps.Get("slug")
	if len(forumSlug) <= 0 {
		http.Error(w, "incorrect slug", http.StatusBadRequest)
		return
	}
	forum, err := db.SelectForumInfo(forumSlug, false)
	if err != nil {
		if len(forum.Slug) > 0 {
			Get404(w, err.Error())
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	js, err := json.Marshal(forum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(js)
}

func forumGetUsers(w http.ResponseWriter, req *http.Request, ps denco.Params) {
	log.Println("forum get users", req.RequestURI)
	slugOrId := ps.Get("slug")
	var err error
	limit := int64(100)
	if limitStr := req.URL.Query().Get("limit"); len(limitStr) > 0 {
		limit, err = strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			limit = 100
		}
	}
	since := req.URL.Query().Get("since")
	if len(since) <= 0 {
		since = ""
	}

	desc, err := strconv.ParseBool(req.URL.Query().Get("desc"))
	if err != nil {
		desc = false
	}

	err = db.SelectForumUsers(slugOrId, int32(limit), since, desc, w)

	if err != nil {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _, _ = json.MarshalToHTTPResponseWriter(models.NotFoundPage{err.Error()}, w)
		return
	}

	//output, err := json.Marshal(users)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//
	//w.Header().Set("content-type", "application/json")
	//_, _ = w.Write(output)
}

func forumGetThreads(w http.ResponseWriter,req *http.Request, ps denco.Params) {
	log.Println("forum get threads:", req.RequestURI)
	slugOrId := ps.Get("slug")
	var err error
	limit := int64(100)
	if limitStr := req.URL.Query().Get("limit"); len(limitStr) > 0 {
		limit, err = strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			limit = 100
		}
	}
	since := req.URL.Query().Get("since")
	if len(since) <= 0 {
		since = ""
	}

	desc, err := strconv.ParseBool(req.URL.Query().Get("desc"))
	if err != nil {
		desc = false
	}

	users, err := db.SelectForumThreads(slugOrId, int32(limit), since, desc)

	if err != nil {
		if users != nil && users[0].Uid < 0 {
			Get404(w, err.Error())
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
}

func forumCreateThread(w http.ResponseWriter,req *http.Request, ps denco.Params) {
	log.Println("forum create thread", req.RequestURI)
	slugOrId := ps.Get("slug")
	data := models.ThreadInfo{}
	err := json.UnmarshalFromReader(req.Body, &data)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	isMin := false
	if data.Slug == nil {
		isMin = true
	}
	thread, err := db.InsertIntoThread(slugOrId, data)
	if err != nil {
		if  err.Error() == "thread exist" {
			output := []byte{}
			if isMin {
				output, err = json.Marshal(models.ThreadInfoMin{
					Uid:     thread.Uid,
					Title:   thread.Title,
					Author:  thread.Author,
					Forum:   thread.Forum,
					Message: thread.Message,
					Created: thread.Created})
			} else {
				output, err = json.Marshal(thread)
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write(output)
			return
		}
		
		if thread.Uid < 0 {
			Get404(w, err.Error())
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	output := []byte{}
	if isMin {
		output, err = json.Marshal(models.ThreadInfoMin{
			Uid:     thread.Uid,
			Title:   thread.Title,
			Author:  thread.Author,
			Forum:   thread.Forum,
			Message: thread.Message,
			Created: thread.Created})
	} else {
		output, err = json.Marshal(thread)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(output)
}


func ForumHandler(router **denco.Mux) []denco.Handler {
	fmt.Println("forums handlers initialized")
	return []denco.Handler{
		(*router).POST("/api/forum/create",        forumCreate),
		(*router).GET( "/api/forum/:slug/details", forumGetInfo),
		(*router).POST("/api/forum/:slug/create",  forumCreateThread),
		(*router).GET( "/api/forum/:slug/users",   forumGetUsers),
		(*router).GET( "/api/forum/:slug/threads", forumGetThreads)}
}

