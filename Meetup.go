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
	"google.golang.org/appengine/user"
	"strconv"
)

const datastoreMeetupsKind = "Meetups"

type Meetup struct {
	Owner         string
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
	GetMeetup(ctx context.Context, id int64) (Meetup, error)
	GetAllMeetups(ctx context.Context) ([]int64, []Meetup, error)
	PutMeetup(ctx context.Context, id int64, meetup *Meetup) error
	AddMeetup(ctx context.Context, meetup *Meetup) (int64, error)
	DeleteMeetup(ctx context.Context, id int64) error
}

// Register meetup routes to the router
func RegisterMeetupRoutes(m *mux.Router, MeetupStorage MeetupStore, PresentationStorage PresentationStore, SpeakerStorage SpeakerStore, MeetupAPIUpdateFunction func() error, MeetupAPICreateFunction func(string) error) error {
	if m == nil {
		return errors.New("m may not be nil when regitering meetup routes")
	}
	h := meetupHandler{MeetupStorage: MeetupStorage, PresentationStorage: PresentationStorage, SpeakerStorage: SpeakerStorage, MeetupAPIUpdateFunction: MeetupAPIUpdateFunction}
	m.HandleFunc("/{ID}/", h.GetMeetup).Methods("GET")
	m.HandleFunc("/", h.AddMeetup).Methods("POST")
	m.HandleFunc("/{id}/delete", h.DeleteMeetup).Methods("GET")
	m.HandleFunc("/{ID}/update", h.UpdateMeetup).Methods("POST")
	m.HandleFunc("/list", h.ListMeetups).Methods("GET")
	m.HandleFunc("/form/add", addMeetupForm).Methods("GET")

	return nil
}

type meetupHandler struct {
	MeetupStorage           MeetupStore
	PresentationStorage     PresentationStore
	SpeakerStorage          SpeakerStore
	MeetupAPIUpdateFunction func() error
	MeetupAPICreateFunction func(string) error
}

func (h *meetupHandler) GetMeetup(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	vars := mux.Vars(r)
	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "ID not valid: %v", vars["ID"])
	}

	meetup, err := h.MeetupStorage.GetMeetup(ctx, ID)
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

func (h *meetupHandler) AddMeetup(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, fmt.Sprint("/meetup/form/add"))
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
	}

	meetup := Meetup{}

	err := json.NewDecoder(r.Body).Decode(&meetup)
	if err != nil {
		log.Errorf(ctx, "Couldn't decode add meetup request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if time.Since(meetup.Date) > time.Second*0 || meetup.Title == "" || time.Since(meetup.VoteTimeEnd) > time.Second*0 || meetup.Description == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Date, title, vote time end and description are mandatory. Date and vote time end need to be in the future.")
		return
	}

	meetup.Owner = u.Email

	ID, err := h.MeetupStorage.AddMeetup(ctx, &meetup)
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", ID)

	err = h.MeetupAPICreateFunction(meetup.Title)
	if err != nil {
		log.Errorf(ctx, "Error when updating meetup API: %v", err)
		return
	}
}

func addMeetupForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Adding Meetup Form</h1>"+
		"<form action=\"/meetup/\" method=\"POST\">"+
		"Title: <input type=\"text\" name=\"Title\"><br>"+
		"Description: <textarea name=\"Description\"></textarea><br>"+
		"<input type=\"submit\" value=\"Save\">"+
		"</form>")
}

func (h *meetupHandler) DeleteMeetup(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	vars := mux.Vars(r)
	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "ID not valid: %v", vars["ID"])
	}

	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, fmt.Sprintf("/meetup/%v/delete", ID))
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
	}

	meetup, err := h.MeetupStorage.GetMeetup(ctx, ID)
	if err == datastore.ErrNoSuchEntity {
		fmt.Fprint(w, "Speaker not found.")
		return
	}
	if err != nil {
		log.Errorf(ctx, "Can't get speaker: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if meetup.Owner != u.Email && !u.Admin {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "You're not the owner nor the admin.")
		return
	}

	err = h.MeetupStorage.DeleteMeetup(ctx, ID)
	if err != nil {
		log.Errorf(ctx, "Can't delete meetup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusTeapot)
	fmt.Fprint(w, "Meetup deleted successfully.")

	err = h.MeetupAPIUpdateFunction()
	if err != nil {
		log.Errorf(ctx, "Error when updating meetup API: %v", err)
		return
	}
}

func (h *meetupHandler) UpdateMeetup(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	vars := mux.Vars(r)
	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "ID not valid: %v", vars["ID"])
	}

	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, fmt.Sprint("/meetup/form/update"))
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
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

	meetup, err := h.MeetupStorage.GetMeetup(ctx, ID)
	if err == datastore.ErrNoSuchEntity {
		fmt.Fprint(w, "No such meetup found.")
		return
	}
	if err != nil {
		log.Errorf(ctx, "Can't get meetup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check if it's the owner
	if meetup.Owner != u.Email && !u.Admin {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "You're not the owner nor the admin.")
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

	err = h.MeetupStorage.PutMeetup(ctx, ID, &meetup)
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "Meetup updated.")

	err = h.MeetupAPIUpdateFunction()
	if err != nil {
		log.Errorf(ctx, "Error when updating meetup API: %v", err)
		return
	}
}

func (h *meetupHandler) ListMeetups(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	keys, meetups, err := h.MeetupStorage.GetAllMeetups(ctx)
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
