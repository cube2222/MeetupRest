package MeetupRest

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Presentation struct {
	Title       string
	Description string
	Speaker     string
	VoteCount   int
}

func GetPresentationHandler() http.Handler {
	m := mux.NewRouter()
	//m.Methods("GET").HandleFunc("/presentation/")
	//m.Methods("POST").HandlerFunc("/presentation/")

	return m
}
