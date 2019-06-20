package handlers

import (
	"../db"
	"../models"
	"fmt"
	json "github.com/mailru/easyjson"
	"github.com/naoina/denco"
	"log"
	"net/http"
	"strconv"
)

func threadChangeInfo(w http.ResponseWriter,req *http.Request, ps denco.Params) {
	slugOrId := ps.Get("slug_or_id")
	thread := models.ThreadInfo{}
	err := json.UnmarshalFromReader(req.Body, &thread)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = db.UpdateThread(slugOrId, &thread)
	if err != nil {
		if thread.Uid == -1 {
			Get404(w, err.Error())
			return
		}
	}

	output, err := json.Marshal(thread)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
}

func threadCreate(w http.ResponseWriter,req *http.Request, ps denco.Params) {
	log.Println("thread create", req.RequestURI)
	slugOrId := ps.Get("slug_or_id")
	data := models.Posts{}
	err := json.UnmarshalFromReader(req.Body, &data)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	forum, err := db.InsertNewPosts(slugOrId, data)
	if err != nil && len(forum) > 0 {
		w.Header().Set("content-type", "application/json")
		if forum[0].Uid == -1 {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusConflict)
		}
		_, _, _ = json.MarshalToHTTPResponseWriter(models.NotFoundPage{err.Error()}, w)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _, _ = json.MarshalToHTTPResponseWriter(forum, w)
}

func threadGetInfo(w http.ResponseWriter,req *http.Request, ps denco.Params) {
	log.Println("thread get info", req.RequestURI)
	slugOrId := ps.Get("slug_or_id")
	_, err := strconv.ParseInt(slugOrId, 10, 64)
	thread := models.ThreadInfo{}
	if err != nil {
		err = db.SelectFromThread(slugOrId, false, &thread)
	} else {
		err = db.SelectFromThread(slugOrId, true, &thread)
	}

	if err != nil {
		if thread.Uid == -1 {
			Get404(w, err.Error())
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output, err := json.Marshal(thread)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
}


func threadGetPosts(w http.ResponseWriter, req *http.Request, ps denco.Params) {
	log.Println("thread get posts:", req.RequestURI)
	slugOrId := ps.Get("slug_or_id")
	var err error
	limit := int64(100)
	if limitStr := req.URL.Query().Get("limit"); len(limitStr) != 0 {
		limit, err = strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			limit = 100
		}
	}

	since := int64(0)
	if sinceStr := req.URL.Query().Get("since"); len(sinceStr) != 0 {
		since, err = strconv.ParseInt(sinceStr, 10, 64)
		if err != nil {
			since = 0
		}
	}
	sort := req.URL.Query().Get("sort")
	if len(sort) == 0 {
		sort = "flat"
	}
	desc, err := strconv.ParseBool(req.URL.Query().Get("desc"))
	if err != nil {
		desc = false
	}

	posts, err := db.SelectThreadPosts(slugOrId, int32(limit), since, sort, desc, w)
	if err != nil {
		if posts != nil {
			if posts[0].Uid == -1 {
				Get404(w, err.Error())
				return
			}
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func threadVote(w http.ResponseWriter,req *http.Request, ps denco.Params) {
	log.Println("thread vote", req.RequestURI)
	slugOrId := ps.Get("slug_or_id")
	voteInfo := models.VoteInfo{}
	err := json.UnmarshalFromReader(req.Body, &voteInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	thread, err := db.UpdateVote(slugOrId, voteInfo)
	if err != nil {
		if thread.Uid == -1 {
			Get404(w, err.Error())
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	threadData := models.FullThreadInfo{
		"author" : thread.Author,
		"created": thread.Created,
		"forum"  : thread.Forum,
		"id"     : thread.Uid,
		"message": thread.Message,
		"slug"   : thread.Slug,
		"title"  : thread.Title,
		"votes"  : thread.Votes,
	}

	output, err := json.Marshal(threadData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
}

func ThreadHandler(router **denco.Mux) []denco.Handler {
	fmt.Println("threads handlers initialized")
	return []denco.Handler{
		(*router).POST("/api/thread/:slug_or_id/create",  threadCreate),
		(*router).GET( "/api/thread/:slug_or_id/details", threadGetInfo),
		(*router).POST("/api/thread/:slug_or_id/details", threadChangeInfo),
		(*router).GET( "/api/thread/:slug_or_id/posts",   threadGetPosts),
		(*router).POST("/api/thread/:slug_or_id/vote",    threadVote)}
}
