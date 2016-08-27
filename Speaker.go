package MeetupRest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

const datastoreSpeakersKind = "Speakers"

type Speaker struct {
	Owner   string
	Name    string
	Surname string
	About   string
	Email   string
	Company string
}

type SpeakerPublicView struct {
	Key     int64
	Name    string
	Surname string
	About   string
	Email   string
	Company string
}

type SpeakerUpdateForm struct {
	NewName    string
	NewSurname string
	NewAbout   string
	NewEmail   string
	NewCompany string
}

// Get the handler which contains all the speaker handling routes and the corresponding handlers.
func RegisterSpeakerRoutes(m *mux.Router) error {
	if m == nil {
		return errors.New("m may not be nil when registering speaker routes")
	}
	m.HandleFunc("/{ID}/", getSpeaker).Methods("GET")
	m.HandleFunc("/", addSpeaker).Methods("POST")
	m.HandleFunc("/list", listSpeakers).Methods("GET")
	m.HandleFunc("/update", updateSpeaker).Methods("POST")
	m.HandleFunc("/{ID}/delete", deleteSpeaker).Methods("GET")
	m.HandleFunc("/form/add", addSpeakerForm).Methods("GET")
	m.HandleFunc("/form/update", updateSpeakerForm).Methods("GET")

	return nil
}

func getSpeaker(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	vars := mux.Vars(r)
	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "ID not valid: %v", vars["ID"])
	}

	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	speaker, err := GetSpeakerByKey(newCtx, ID)
	done()
	if err == datastore.ErrNoSuchEntity {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Couldn't find speaker with id: %v", ID)
		return
	}
	if err != nil {
		log.Errorf(ctx, "Couldn't get speaker: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	speakerPublicView := speaker.GetPublicView(ID)
	err = speakerPublicView.WriteTo(w)
	if err != nil {
		log.Errorf(ctx, "Failed to write speaker: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func addSpeaker(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, fmt.Sprint("/speaker/form/add"))
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
	}

	err := r.ParseForm()
	if err != nil {
		log.Errorf(ctx, "Couldn't parse form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s := &Speaker{}

	decoder := schema.NewDecoder()
	decoder.Decode(s, r.PostForm)

	if s.Name == "" || s.Surname == "" || s.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Name, surname and email are mandatory.")
		return
	}

	s.Owner = u.Email

	key := datastore.NewKey(ctx, datastoreSpeakersKind, "", 0, nil)
	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	id, err := datastore.Put(newCtx, key, s)
	done()
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", id.IntID())
}

func addSpeakerForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Adding Speaker Form</h1>"+
		"<form action=\"/speaker/\" method=\"POST\">"+
		"First name: <input type=\"text\" name=\"Name\"><br>"+
		"Last name: <input type=\"text\" name=\"Surname\"><br>"+
		"Company: <input type=\"text\" name=\"Company\"><br>"+
		"Email: <input type=\"email\" name=\"Email\"><br>"+
		"About: <textarea name=\"About\"></textarea><br>"+
		"<input type=\"submit\" value=\"Save\">"+
		"</form>")
}

func updateSpeaker(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	vars := mux.Vars(r)
	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "ID not valid: %v", vars["ID"])
	}

	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, fmt.Sprint("/speaker/form/update"))
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Errorf(ctx, "Couldn't parse form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	suf := &SpeakerUpdateForm{}
	decoder := schema.NewDecoder()
	decoder.Decode(suf, r.PostForm)

	k := datastore.NewKey(ctx, datastoreSpeakersKind, "", ID, nil)

	mySpeaker := &Speaker{}

	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	err = datastore.Get(newCtx, k, &mySpeaker)
	done()

	if err == datastore.ErrNoSuchEntity {
		fmt.Fprint(w, "No speaker with ID: %v", ID)
		return
	}
	// Some other error
	if err != nil {
		log.Errorf(ctx, "Can't get speaker: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Check if it's the owner
	if mySpeaker.Owner != u.Email && !u.Admin {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "You're not the owner nor the admin.")
		return
	}
	//TODO: Update speaker with function.
	if suf.NewName != "" {
		mySpeaker.Name = suf.NewName
	}

	if suf.NewSurname != "" {
		mySpeaker.Surname = suf.NewSurname
	}

	if suf.NewEmail != "" {
		mySpeaker.Email = suf.NewEmail
	}

	if suf.NewCompany != "" {
		mySpeaker.Company = suf.NewCompany
	}

	if suf.NewAbout != "" {
		mySpeaker.About = suf.NewAbout
	}

	newCtx, done = context.WithTimeout(ctx, time.Second*2)
	_, err = datastore.Put(newCtx, k, mySpeaker)
	done()
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "Speaker updated.")
}

func updateSpeakerForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Update Speaker Form</h1>"+
		"<form action=\"/speaker/update\" method=\"POST\">"+
		"Current First name: <input type=\"text\" name=\"CurrentName\"><br>"+
		"Current Last name: <input type=\"text\" name=\"CurrentSurname\"><br>"+
		"Current Email: <input type=\"email\" name=\"CurrentEmail\"><br>"+
		"Leave fields not to update blank:<br>"+
		"New first name: <input type=\"text\" name=\"NewName\"><br>"+
		"New last name: <input type=\"text\" name=\"NewSurname\"><br>"+
		"New company: <input type=\"text\" name=\"NewCompany\"><br>"+
		"New email: <input type=\"email\" name=\"NewEmail\"><br>"+
		"New about: <textarea name=\"NewAbout\"></textarea><br>"+
		"<input type=\"submit\" value=\"Save\">"+
		"</form>")
}

func deleteSpeaker(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	vars := mux.Vars(r)
	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "ID not valid: %v", vars["ID"])
	}

	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, fmt.Sprintf("/speaker/%v/delete", ID))
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
	}

	mySpeaker := Speaker{}

	k := datastore.NewKey(ctx, datastoreSpeakersKind, "", ID, nil)
	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	err = datastore.Get(newCtx, k, &mySpeaker)
	done()
	if err == datastore.ErrNoSuchEntity {
		fmt.Fprint(w, "Speaker not found.")
		return
	}
	if err != nil {
		log.Errorf(ctx, "Can't get speaker: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if mySpeaker.Owner != u.Email && !u.Admin {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "You're not the owner nor the admin.")
		return
	}

	newCtx, done = context.WithTimeout(ctx, time.Second*2)
	err = datastore.Delete(newCtx, k)
	done()
	if err != nil {
		log.Errorf(ctx, "Can't delete speaker: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusTeapot)
	fmt.Fprint(w, "Speaker deleted successfully.")
}

func listSpeakers(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	speakers := make([]Speaker, 0, 10)

	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	keys, err := datastore.NewQuery(datastoreSpeakersKind).GetAll(newCtx, &speakers)
	done()
	if err != nil {
		log.Errorf(ctx, "Can't get speakers: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	speakersPublicView := make([]SpeakerPublicView, 0, len(speakers))
	for index, speaker := range speakers {
		speakersPublicView = append(speakersPublicView, speaker.GetPublicView(keys[index].IntID()))
	}

	err = WriteSpeakersPublicView(speakersPublicView, w)
	if err != nil {
		log.Errorf(ctx, "Failed to write speakers slice: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetSpeakerByKey(ctx context.Context, key int64) (Speaker, error) {
	speaker := Speaker{}
	speakerKey := datastore.NewKey(ctx, datastoreSpeakersKind, "", key, nil)
	err := datastore.Get(ctx, speakerKey, &speaker)
	return speaker, err
}

func (speaker *Speaker) GetSpeakerFullName() string {
	return fmt.Sprintf("%v %v", speaker.Name, speaker.Surname)
}

func (s *Speaker) GetPublicView(key int64) SpeakerPublicView {
	return SpeakerPublicView{
		Key:     key,
		Name:    s.Name,
		Surname: s.Surname,
		About:   s.About,
		Email:   s.Email,
		Company: s.Company,
	}
}

func (s *Speaker) WriteTo(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(s)
}

func (s *SpeakerPublicView) WriteTo(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(s)
}

func WriteSpeakersPublicView(speakers []SpeakerPublicView, w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(speakers)
}
