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

const kindMeetups = "Meetup"

type Meetup struct {
	Title         string
	Description   string
	Presentations []string
	Date          time.Time
	VoteTimeEnd   time.Time
}

func GetMeetupHandler() http.Handler {
	m := mux.NewRouter()
	m.HandleFunc("/meetup", getMeetup).Methods("GET")
	m.HandleFunc("/meetup", addMeetup).Methods("POST")

	return m
}

func getMeetup(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Errorf(ctx, "Can't parse query: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Can't parse query: %v", err)
		return
	}

	q := datastore.NewQuery(kindMeetups).Limit(1)

	if title, ok := params["title"]; ok == true {
		q = q.Filter("Title=", title[0])
	}

	if description, ok := params["description"]; ok == true {
		q = q.Filter("Description=", description[0])
	}

	if presentation, ok := params["presentation"]; ok == true {
		//TODO: Add the ability o add tables
		q = q.Filter("Presentations=", presentation[0])
	}

	if date, ok := params["date"]; ok == true {
		q = q.Filter("Date=", date[0])
	}

	t := q.Run(ctx)
	myMeetup := Meetup{}
	_, err = t.Next(&myMeetup)
	if err == datastore.Done {
		fmt.Fprint(w, "No meetup found.")
		return
	}
	// Some other error.
	if err != nil {
		log.Errorf(ctx, "Can't get speaker: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Can't get speaker: %v", err)
		return
	}
	data, err := json.Marshal(&myMeetup)
	if err != nil {
		log.Errorf(ctx, "Failed to serialize speaker: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to serialize speaker: %v", err)
		return
	}
	io.Copy(w, bytes.NewBuffer(data))
}

func addMeetup(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	err := r.ParseForm()
	if err != nil {
		log.Errorf(ctx, "Couldn't parse form: Meetup")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Couldn't parse form: Meetup")
		return
	}

	m := &Meetup{}

	decoder := schema.NewDecoder()
	decoder.Decode(m, r.PostForm)

	if /*m.Date == nil ||*/ m.Title == "" || /*m.VoteTimeEnd == nil ||*/ m.Description == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Date, tile vote time end and description are mandatory.")
		return
	}

	key := datastore.NewKey(ctx, kindMeetups, "", 0, nil)
	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	id, err := datastore.Put(newCtx, key, m)
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Can't create datastore object: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", id.IntID())
}
