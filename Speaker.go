package MeetupRest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Speaker struct {
	Name    string
	Surname string
	About   string
	Email   string
	Company string
}

func GetSpeakerHandler() http.Handler {
	m := mux.NewRouter()
	m.HandleFunc("/speaker", getSpeaker).Methods("GET")
	m.HandleFunc("/speaker", addSpeaker).Methods("POST")

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

	q := datastore.NewQuery("Speaker") //.Limit(1)

	if name, ok := params["name"]; ok == true {
		q = q.Filter("Name=", name)
	}

	if surname, ok := params["surname"]; ok == true {
		q = q.Filter("Surname=", surname)
	}

	if email, ok := params["email"]; ok == true {
		q = q.Filter("Email=", email)
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

func addSpeaker(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	err := r.ParseForm()
	if err != nil {
		log.Errorf(ctx, "Couldn't parse form: Name")
		// TODO: Status Code
		fmt.Fprint(w, "Couldn't parse form: Name")
		return
	}

	s := &Speaker{}

	decoder := schema.NewDecoder()
	decoder.Decode(s, r.PostForm)

	if s.Name == "" || s.Surname == "" || s.Email == "" {
		// TODO: Status Code
		fmt.Fprint(w, "Name, surname and email are mandatory.")
		return
	}

	key := datastore.NewKey(ctx, "Speaker", "", 0, nil)
	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	id, err := datastore.Put(newCtx, key, s)
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", id.IntID())
}
