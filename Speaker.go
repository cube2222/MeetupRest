package MeetupRest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"io"
	"net/http"
	"net/url"
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
	m.Methods("GET").HandlerFunc("/speaker/", getSpeaker)
	m.Methods("POST").HandlerFunc("/speaker/", addSpeaker)

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

func addSpeaker(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Errorf(ctx, "Can't parse query: %v", err)
		// TODO: http code
		fmt.Fprintf(w, "Can't parse query: %v", err)
		return
	}

	s := &Speaker{}

	if s.Name, ok = params.Get("Name"); ok == false {
		log.Error(ctx, "Missing parametr: Name")
		fmt.Fprint(w, "Missing parametr: Name")
		return
	}

	if s.Surname, ok = params.Get("Surname"); ok == false {
		log.Error(ctx, "Missing parametr: Surname")
		fmt.Fprint(w, "Missing parametr: Surname")
		return
	}

	if s.Email, ok = params.Get("Email"); ok == false {
		log.Error(ctx, "Missing parametr: Email")
		fmt.Fprint(w, "Missing parametr: Email")
		return
	}

	s.Company = params.Get("Company")
	s.About = params.Get("About")

	key := datastore.NewKey(ctx, "Speaker", "", 0, nil)
	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	id, err := datastore.Put(newCtx, key, s)
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
