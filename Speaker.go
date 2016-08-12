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

const datastoreSpeakersKind = "Speakers"

type Speaker struct {
	Name    string
	Surname string
	About   string
	Email   string
	Company string
}

type SpeakerUpdateForm struct {
	CurrentName    string
	CurrentSurname string
	CurrentEmail   string
	NewName        string
	NewSurname     string
	NewAbout       string
	NewEmail       string
	NewCompany     string
}

// Get the handler which contains all the speaker handling routes and the corresponding handlers.
func RegisterSpeakerRoutes(m *mux.Router) error {
	if m == nil {
		return errors.New("m may not be nil when registering speaker routes")
	}
	m.HandleFunc("/", getSpeaker).Methods("GET")
	m.HandleFunc("/", addSpeaker).Methods("POST")
	m.HandleFunc("/update", updateSpeaker).Methods("POST")
	m.HandleFunc("/form/add", addSpeakerForm).Methods("GET")
	m.HandleFunc("/form/update", updateSpeakerForm).Methods("GET")

	return nil
}

func getSpeaker(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Errorf(ctx, "Can't parse query: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	q := datastore.NewQuery(datastoreSpeakersKind).Limit(1)

	if name, ok := params["name"]; ok == true {
		q = q.Filter("Name=", name[0])
	}

	if surname, ok := params["surname"]; ok == true {
		q = q.Filter("Surname=", surname[0])
	}

	if email, ok := params["email"]; ok == true {
		q = q.Filter("Email=", email[0])
	}

	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	t := q.Run(newCtx)

	mySpeaker := Speaker{}
	_, err = t.Next(&mySpeaker)
	// No speaker retrieved
	if err == datastore.Done {
		fmt.Fprint(w, "No speaker found.")
		return
	}
	// Some other error.
	if err != nil {
		log.Errorf(ctx, "Can't get speaker: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(&mySpeaker)
	if err != nil {
		log.Errorf(ctx, "Failed to serialize speaker: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	io.Copy(w, bytes.NewReader(data))
}

func addSpeaker(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

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

	key := datastore.NewKey(ctx, datastoreSpeakersKind, "", 0, nil)
	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	id, err := datastore.Put(newCtx, key, s)
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

	err := r.ParseForm()
	if err != nil {
		log.Errorf(ctx, "Couldn't parse form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	suf := &SpeakerUpdateForm{}
	decoder := schema.NewDecoder()
	decoder.Decode(suf, r.PostForm)

	if suf.CurrentEmail == "" && (suf.CurrentName == "" || suf.CurrentSurname == "") {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Current speaker email or name with surname is mandatory.")
		return
	}
	q := datastore.NewQuery(datastoreSpeakersKind).Limit(1)

	if suf.CurrentEmail != "" {
		q = q.Filter("Email=", suf.CurrentEmail)
	}

	if suf.CurrentName != "" {
		q = q.Filter("Name=", suf.CurrentName)
	}

	if suf.CurrentSurname != "" {
		q = q.Filter("Surname=", suf.CurrentSurname)
	}

	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	t := q.Run(newCtx)

	mySpeaker := &Speaker{}
	key, err := t.Next(&mySpeaker)

	if err == datastore.Done {
		fmt.Fprint(w, "No such speaker found.")
		return
	}
	// Some other error
	if err != nil {
		log.Errorf(ctx, "Can't get speaker: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

	newCtx, _ = context.WithTimeout(ctx, time.Second*2)
	err = datastore.Delete(newCtx, key)
	if err != nil {
		log.Errorf(ctx, "Can't delete datastore object: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	key = datastore.NewKey(ctx, datastoreSpeakersKind, "", 0, nil)

	newCtx, _ = context.WithTimeout(ctx, time.Second*2)
	_, err = datastore.Put(newCtx, key, mySpeaker)
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
