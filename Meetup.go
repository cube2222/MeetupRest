package MeetupRest

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"strconv"
)

const datastoreMeetupsKind = "Meetups"

type Meetup struct {
	Title         string
	Description   string
	Presentations []int64
	Date          time.Time
	VoteTimeEnd   time.Time
}

type MeetupPublicView struct {
	Key           int64
	Title         string
	Description   string
	Presentations []int64
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

type MeetupStore interface {
	GetMeetup(ctx context.Context, id int64) (Meetup, error)
	GetAllMeetups(ctx context.Context) ([]int64, []Meetup, error)
	PutMeetup(ctx context.Context, id int64, meetup *Meetup) error
	AddMeetup(ctx context.Context, meetup *Meetup) (int64, error)
	DeleteMeetup(ctx context.Context, id int64) error
}

// Register meetup routes to the router
func RegisterMeetupRoutes(m *mux.Router, Storage *SpeakerStore) error {
	if m == nil {
		return errors.New("m may not be nil when regitering meetup routes")
	}
	h := meetupHandler{Storage: Storage}
	m.HandleFunc("/{ID}/", h.getMeetup).Methods("GET")
	m.HandleFunc("/", h.addMeetup).Methods("POST")
	m.HandleFunc("/delete", h.deleteMeetup).Methods("DELETE")
	m.HandleFunc("/update", h.updateMeetup).Methods("POST")
	m.HandleFunc("/list", h.listMeetups).Methods("GET")
	m.HandleFunc("/form/add", addMeetupForm).Methods("GET")

	return nil
}

type meetupHandler struct {
	Storage *MeetupStore
}

func (h *meetupHandler) getMeetup(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	vars := mux.Vars(r)
	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "ID not valid: %v", vars["ID"])
	}

	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	meetup, err := GetMeetupByKey(newCtx, ID)
	done()
	if err == datastore.ErrNoSuchEntity {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Couldn't find meetup with id: %v", ID)
		return
	}
	if err != nil {
		log.Errorf(ctx, "Couldn't get meetup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	meetupPublicView := meetup.GetPublicView(ID)
	err = meetupPublicView.WriteTo(w)
	if err != nil {
		log.Errorf(ctx, "Failed to write meetup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *meetupHandler) addMeetup(w http.ResponseWriter, r *http.Request) {
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
	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	id, err := datastore.Put(newCtx, key, m)
	done()
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", id.IntID())
}

func addMeetupForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Adding Meetup Form</h1>"+
		"<form action=\"/meetup/\" method=\"POST\">"+
		"Title: <input type=\"text\" name=\"Title\"><br>"+
		"Description: <textarea name=\"Description\"></textarea><br>"+
		"<input type=\"submit\" value=\"Save\">"+
		"</form>")
}

func (h *meetupHandler) deleteMeetup(w http.ResponseWriter, r *http.Request) {
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

	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	t := q.Run(newCtx)
	done()
	myMeetup := Meetup{}
	key, err := t.Next(&myMeetup)
	if err == datastore.Done {
		fmt.Fprint(w, "No meetup found.")
		return
	}

	newCtx, done = context.WithTimeout(ctx, time.Second*2)
	err = datastore.Delete(newCtx, key)
	done()
	if err != nil {
		log.Errorf(ctx, "Can't delete meetup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusTeapot)
	fmt.Fprint(w, "Meetup deleted successfully.")
}

func (h *meetupHandler) updateMeetup(w http.ResponseWriter, r *http.Request) {
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

	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	t := datastore.NewQuery(datastoreMeetupsKind).
		Filter("Title=", muf.CurrentTitle).
		Limit(1).
		Run(newCtx)
	done()
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

	newCtx, done = context.WithTimeout(ctx, time.Second*2)
	_, err = datastore.Put(newCtx, key, myMeetup)
	done()
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "Meetup updated.")
}

func (h *meetupHandler) listMeetups(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	meetups := make([]Meetup, 0, 10)

	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	keys, err := datastore.NewQuery(datastoreMeetupsKind).GetAll(newCtx, &meetups)
	done()
	if err != nil {
		log.Errorf(ctx, "Can't get meetups: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	meetupsPublicView := make([]MeetupPublicView, 0, len(meetups))
	for index, meetup := range meetups {
		meetupsPublicView = append(meetupsPublicView, meetup.GetPublicView(keys[index].IntID()))
	}

	err = WriteMeetupPublicView(meetupsPublicView, w)
	if err != nil {
		log.Errorf(ctx, "Failed to write meetups slice: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetMeetupByKey(ctx context.Context, key int64) (Meetup, error) {
	meetup := Meetup{}
	meetupKey := datastore.NewKey(ctx, datastoreMeetupsKind, "", key, nil)
	err := datastore.Get(ctx, meetupKey, &meetup)
	return meetup, err
}

func (m *Meetup) GetPublicView(key int64) MeetupPublicView {
	return MeetupPublicView{
		Key:           key,
		Title:         m.Title,
		Description:   m.Description,
		Presentations: m.Presentations,
		Date:          m.Date,
		VoteTimeEnd:   m.VoteTimeEnd,
	}
}

func (m *Meetup) WriteTo(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(m)
}

func (m *MeetupPublicView) WriteTo(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(m)
}

func WriteMeetupPublicView(meetups []MeetupPublicView, w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(meetups)
}
