package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db            *sql.DB
	err           error
	authenticated = false
)

func main() {
	db, err = sql.Open("sqlite3", "db.sqlite3")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	// test connection
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	os.Setenv("PORT", "2300")
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}
	// route
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/forgot", forgotHandler)
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.Handle("/css/",
		http.StripPrefix("/css/", http.FileServer(http.Dir("./css"))),
	)
	http.ListenAndServe(":"+port, nil)
}
