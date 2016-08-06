package MeetupRest

import (
	"net/http"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"net/url"
	"google.golang.org/appengine/log"
	"fmt"
	"google.golang.org/appengine/datastore"
	"encoding/json"
	"io"
	"bytes"
)

type Speaker struct {
	Name          string
	Surname       string
	About         string
	Email         string
	Company       string
	Presentations []string
}

func GetSpeakerHandler() http.Handler {
	m := mux.NewRouter()
	m.HandleFunc(/...)
	m.Methods("GET").HandleFunc("/author/")

	return m
}

func getSpeaker(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Errorf(ctx, "Can't parse query: %v", err)
		// TODO: http code
		fmt.Fprintf(w, "Can't parse query: %v", err)
		return
	}

	q := datastore.NewQuery("Speaker").Limit(1)

	if name, ok := params["Name"]; ok == true {
		q = q.Filter("Name =", name)
	}
	
	if surname, ok := params["Surname"]; ok == true {
		q = q.Filter("Surname =", surname)
	}

	if email, ok := params["Email"]; ok == true {
		q = q.Filter("Email =", email)
	}

	t := q.Run(ctx)

	mySpeaker := Speaker{}
	_, err = t.Next(&mySpeaker)
	if err == datastore.Done {
		fmt.Fprint(w, "No speaker found.")
		return
	}
	if err != nil {
		log.Errorf(ctx, "Can't get speaker: %v", err)
		// TODO: http code
		fmt.Fprintf(w, "Can't get speaker: %v", err)
		return
	}
	data, err := json.Marshal(&mySpeaker)
	if err != nil {
		log.Errorf(ctx, "Failed to serialize speaker: %v", err)
		// TODO: http code
		fmt.Fprintf(w, "Failed to serialize speaker: %v", err)
		return
	}
	io.Copy(w, bytes.NewBuffer(data))
}
