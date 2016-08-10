package MeetupRest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const datastoreMeetupsKind = "Meetups"

type Meetup struct {
	Title         string
	Description   string
	Presentations []string
	Date          time.Time
	VoteTimeEnd   time.Time
}

// Register meetup routes to the router
func RegisterMeetupRoutes(m *mux.Router) error {
	if m == nil {
		return errors.New("m may not be nil when regitering meetup routes")
	}
	m.HandleFunc("/", getMeetup).Methods("GET")
	m.HandleFunc("/", addMeetup).Methods("POST")
	m.HandleFunc("/list", getAllMeetups).Methods("GET")

	return nil
}

func getMeetup(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	log.Infof(ctx, "Received meetup get.")

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Errorf(ctx, "Can't parse query: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Can't parse query: %v", err)
		return
	}

	q := datastore.NewQuery(datastoreMeetupsKind).Limit(1)

	if title, ok := params["title"]; ok == true {
		q = q.Filter("Title=", title[0])
	}

	if description, ok := params["description"]; ok == true {
		q = q.Filter("Description=", description[0])
	}

	if presentation, ok := params["presentation"]; ok == true {
		//TODO: Add the ability to add tables
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
		log.Errorf(ctx, "Can't get meetup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Can't get meetup: %v", err)
		return
	}
	data, err := json.Marshal(&myMeetup)
	if err != nil {
		log.Errorf(ctx, "Failed to serialize meetup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to serialize meetup: %v", err)
		return
	}
	io.Copy(w, bytes.NewBuffer(data))
}

func getAllMeetups(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	log.Infof(ctx, "Received meetup list.")

	meetups := make([]Meetup, 0, 10)
	_, err := datastore.NewQuery(datastoreMeetupsKind).GetAll(ctx, &meetups)
	if err != nil {
		log.Errorf(ctx, "Can't get meetups: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Can't get meetups: %v", err)
		return
	}

	data, err := json.Marshal(&meetups)
	if err != nil {
		log.Errorf(ctx, "Failed to serialize meetups array(slice): %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to serialize meetups array(slice) : %v", err)
		return
	}
	io.Copy(w, bytes.NewBuffer(data))
}

func addMeetup(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	log.Infof(ctx, "Received meetup post.")

	err := r.ParseForm()
	if err != nil {
		log.Errorf(ctx, "Couldn't parse form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Couldn't parse form: %v", err)
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

	key := datastore.NewKey(ctx, datastoreMeetupsKind, "", 0, nil)
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
