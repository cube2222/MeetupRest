package MeetupRest

import (
	"net/http"
	"github.com/gorilla/mux"
)

type Author struct {
	Name          string
	Surname       string
	About         string
	Email         string
	Company       string
	Presentations []string
}

func GetAuthorHandler() http.Handler {
	m := mux.NewRouter()
	m.HandleFunc(/...)
	m.Methods("GET").HandleFunc("/author/")

	return m
}