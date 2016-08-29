package MeetupRest

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"io"
	"net/http"
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
	NewTitle       string
	NewDescription string
	NewDate        time.Time
	NewVoteTimeEnd time.Time
}

type MeetupStore interface {
	PresentationStore
	GetMeetup(ctx context.Context, id int64) (Meetup, error)
	GetAllMeetups(ctx context.Context) ([]int64, []Meetup, error)
	PutMeetup(ctx context.Context, id int64, meetup *Meetup) error
	AddMeetup(ctx context.Context, meetup *Meetup) (int64, error)
	DeleteMeetup(ctx context.Context, id int64) error
}

// Register meetup routes to the router
func RegisterMeetupRoutes(m *mux.Router, Storage MeetupStore) error {
	if m == nil {
		return errors.New("m may not be nil when regitering meetup routes")
	}
	h := meetupHandler{Storage: Storage}
	m.HandleFunc("/{ID}/", h.getMeetup).Methods("GET")
	m.HandleFunc("/", h.addMeetup).Methods("POST")
	m.HandleFunc("/{id}/delete", h.deleteMeetup).Methods("GET")
	m.HandleFunc("/{ID}/update", h.updateMeetup).Methods("POST")
	m.HandleFunc("/list", h.listMeetups).Methods("GET")
	m.HandleFunc("/form/add", addMeetupForm).Methods("GET")

	return nil
}

type meetupHandler struct {
	Storage MeetupStore
}

func (h *meetupHandler) getMeetup(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	vars := mux.Vars(r)
	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "ID not valid: %v", vars["ID"])
	}

	meetup, err := h.Storage.GetMeetup(ctx, ID)
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
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	err := r.ParseForm()
	if err != nil {
		log.Errorf(ctx, "Couldn't parse form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m := Meetup{}

	decoder := schema.NewDecoder()
	decoder.Decode(&m, r.PostForm)
	if /*m.Date == nil ||*/ m.Title == "" || /*m.VoteTimeEnd == nil ||*/ m.Description == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Date, tile vote time end and description are mandatory.")
		return
	}

	ID, err := h.Storage.AddMeetup(ctx, &m)
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", ID)
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
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	vars := mux.Vars(r)
	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "ID not valid: %v", vars["ID"])
	}

	err = h.Storage.DeleteMeetup(ctx, ID)
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
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	vars := mux.Vars(r)
	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "ID not valid: %v", vars["ID"])
	}

	err = r.ParseForm()
	if err != nil {
		log.Errorf(ctx, "Couldn't parse form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	muf := &MeetupUpdateForm{}
	decoder := schema.NewDecoder()
	decoder.Decode(muf, r.PostForm)

	meetup, err := h.Storage.GetMeetup(ctx, ID)
	if err == datastore.ErrNoSuchEntity {
		fmt.Fprint(w, "No such meetup found.")
		return
	}
	if err != nil {
		log.Errorf(ctx, "Can't get meetup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if muf.NewTitle != "" {
		meetup.Title = muf.NewTitle
	}

	if muf.NewDescription != "" {
		meetup.Description = muf.NewDescription
	}

	/*if muf.NewDate != nil {

	}

	if muf.NewVoteTimeEnd != nil {

	}*/

	err = h.Storage.PutMeetup(ctx, ID, &meetup)
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
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	keys, meetups, err := h.Storage.GetAllMeetups(ctx)
	if err != nil {
		log.Errorf(ctx, "Can't get meetups: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	meetupsPublicView := make([]MeetupPublicView, 0, len(meetups))
	for index, meetup := range meetups {
		meetupsPublicView = append(meetupsPublicView, meetup.GetPublicView(keys[index]))
	}

	err = WriteMeetupPublicView(meetupsPublicView, w)
	if err != nil {
		log.Errorf(ctx, "Failed to write meetups slice: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
