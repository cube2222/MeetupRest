package MeetupRest

import (
	"net/http"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"net/url"
	"google.golang.org/appengine/log"
	"fmt"
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

func getAuthorHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Errorf(ctx, "Can't parse query: %v", err)
		fmt.Fprintf(w, "Can't parse query: %v", err)
		return
	}
	if name, ok := params["name"]; ok == true {

	}
}
