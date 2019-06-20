package main

import (
	"./db"
	"./handlers"
	"github.com/jackc/pgx"
	"github.com/naoina/denco"
	"log"
	"net/http"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "park"
	password = "admin"
	dbname   = "park_forum"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	pgxConfig := pgx.ConnConfig{
		Host:     host,
		Port:     port,
		Database: dbname,
		User:     user,
		Password: password,
	}
	pgxConnPoolConfig := pgx.ConnPoolConfig{
		ConnConfig: pgxConfig,
	}
	var err error
	db.DB, err = pgx.NewConnPool(pgxConnPoolConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	router := denco.NewMux()
	handlerArray := handlers.UserHandler(&router)
	for _, elem := range handlers.ForumHandler(&router) {
		handlerArray = append(handlerArray, elem)
	}
	for _, elem := range handlers.PostHandler(&router) {
		handlerArray = append(handlerArray, elem)
	}
	for _, elem := range handlers.ServiceHandler(&router) {
		handlerArray = append(handlerArray, elem)
	}
	for _, elem := range handlers.ThreadHandler(&router) {
		handlerArray = append(handlerArray, elem)
	}

	handler, err := router.Build(handlerArray)
	http.Handle("/", handler)

	err = http.ListenAndServe(":5000", handler)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
