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

type MeetupUpdateForm struct {
	CurrentTitle   string
	NewTitle       string
	NewDescription string
	NewDate        time.Time
	NewVoteTimeEnd time.Time
}

// Register meetup routes to the router
func RegisterMeetupRoutes(m *mux.Router) error {
	if m == nil {
		return errors.New("m may not be nil when regitering meetup routes")
	}
	m.HandleFunc("/", getMeetup).Methods("GET")
	m.HandleFunc("/", addMeetup).Methods("POST")
	m.HandleFunc("/delete", deleteMeetup).Methods("DELETE")
	m.HandleFunc("/update", updateMeetup).Methods("POST")
	m.HandleFunc("/list", getAllMeetups).Methods("GET")

	return nil
}

func getMeetup(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Errorf(ctx, "Can't parse query: %v", err)
		w.WriteHeader(http.StatusBadRequest)
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

	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	t := q.Run(newCtx)
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
		return
	}
	data, err := json.Marshal(&myMeetup)
	if err != nil {
		log.Errorf(ctx, "Failed to serialize meetup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	io.Copy(w, bytes.NewReader(data))
}

func getAllMeetups(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	meetups := make([]Meetup, 0, 10)
	_, err := datastore.NewQuery(datastoreMeetupsKind).GetAll(newCtx, &meetups)
	if err != nil {
		log.Errorf(ctx, "Can't get meetups: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(&meetups)
	if err != nil {
		log.Errorf(ctx, "Failed to serialize meetups array(slice): %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	io.Copy(w, bytes.NewReader(data))
}

func addMeetup(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	err := r.ParseForm()
	if err != nil {
		log.Errorf(ctx, "Couldn't parse form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
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
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", id.IntID())
}

func deleteMeetup(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Errorf(ctx, "Can't parse query: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	q := datastore.NewQuery(datastoreMeetupsKind).Limit(1)

	title, okTitle := params["title"]
	ID, okID := params["id"]
	if !(okID || okTitle) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Meetup title or ID must be provided.")
		return
	}

	if okTitle {
		q = q.Filter("Title=", title[0])
	}

	if okID {
		q = q.Filter("ID=", ID[0])
	}

	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	t := q.Run(newCtx)
	myMeetup := Meetup{}
	key, err := t.Next(&myMeetup)
	if err == datastore.Done {
		fmt.Fprint(w, "No meetup found.")
		return
	}

	newCtx, _ = context.WithTimeout(ctx, time.Second*2)
	if err = datastore.Delete(newCtx, key); err != nil {
		log.Errorf(ctx, "Can't delete meetup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusTeapot)
	fmt.Fprint(w, "Meetup deleted successfully.")
}

func updateMeetup(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	err := r.ParseForm()
	if err != nil {
		log.Errorf(ctx, "Couldn't parse form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	muf := &MeetupUpdateForm{}
	decoder := schema.NewDecoder()
	decoder.Decode(muf, r.PostForm)

	if muf.CurrentTitle == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Current meetup title is mandatory.")
		return
	}

	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	t := datastore.NewQuery(datastoreMeetupsKind).
		Filter("Title=", muf.CurrentTitle).
		Limit(1).
		Run(newCtx)
	myMeetup := &Meetup{}
	key, err := t.Next(myMeetup)

	if err == datastore.Done {
		fmt.Fprint(w, "No such meetup found.")
		return
	}
	// Some other error
	if err != nil {
		log.Errorf(ctx, "Can't get meetup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if muf.NewTitle != "" {
		myMeetup.Title = muf.NewTitle
	}

	if muf.NewDescription != "" {
		myMeetup.Description = muf.NewDescription
	}

	/*if muf.NewDate != nil {

	}

	if muf.NewVoteTimeEnd != nil {

	}*/

	newCtx, _ = context.WithTimeout(ctx, time.Second*2)
	_, err = datastore.Put(newCtx, key, myMeetup)
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "Meetup updated.")
}
