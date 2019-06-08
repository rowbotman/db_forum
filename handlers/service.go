package handlers

import (
	"../db"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func serviceDrop(w http.ResponseWriter,req *http.Request) {
	w.Header().Set("content-type", "text/plain")
	if db.ClearService() {
		_, _ = w.Write([]byte("Отчистка базы успешно завершена"))
		return
	}
	_, _ = w.Write([]byte("error occurred"))
}

func serviceGetInfo(w http.ResponseWriter,req *http.Request) {
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

func ServiceHandler(router **mux.Router) {
	fmt.Println("services handlers initialized")
	(*router).HandleFunc("/api/service/clear", serviceDrop).Methods("POST")
	(*router).HandleFunc("/api/service/status", serviceGetInfo).Methods("GET")
}
