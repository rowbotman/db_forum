package handlers

import (
	"../db"
	"fmt"
	htmux "github.com/dimfeld/httptreemux"
	json "github.com/mailru/easyjson"
	//"log"
	"net/http"
)

func serviceDrop(w http.ResponseWriter, req *http.Request, _ map[string]string) {
	//log.Println("service drop", req.RequestURI)
	w.Header().Set("content-type", "text/plain")
	if db.ClearService() {
		_, _ = w.Write([]byte("Отчистка базы успешно завершена"))
		return
	}
	_, _ = w.Write([]byte("error occurred"))
}

func serviceGetInfo(w http.ResponseWriter, req *http.Request, _ map[string]string) {
	//log.Println("service get info", req.RequestURI)
	w.Header().Set("content-type", "text/plain")
	status, err := db.ServiceGet()
	if err != nil {
		return
	}

	output, err := json.Marshal(status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
}

func ServiceHandler(router *htmux.TreeMux) {
	fmt.Println("services handlers initialized")
	router.POST("/api/service/clear",  serviceDrop)
	router.GET( "/api/service/status", serviceGetInfo)
}
