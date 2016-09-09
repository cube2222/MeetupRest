package MeetupRest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/context"

	"strconv"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
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

type SpeakerForm struct {
	Name    string
	Surname string
	About   string
	Email   string
	Company string
}

type SpeakerStore interface {
	GetSpeaker(ctx context.Context, id int64) (Speaker, error)
	GetAllSpeakers(ctx context.Context) ([]int64, []Speaker, error)
	PutSpeaker(ctx context.Context, id int64, speaker *Speaker) error
	AddSpeaker(ctx context.Context, speaker *Speaker) (int64, error)
	DeleteSpeaker(ctx context.Context, id int64) error
	GetSpeakerIdByName(ctx context.Context, name string) (int64, error)
}

// Get the handler which contains all the speaker handling routes and the corresponding handlers.
func RegisterSpeakerRoutes(m *mux.Router, SpeakerStorage SpeakerStore) error {
	if m == nil {
		return errors.New("m may not be nil when registering speaker routes")
	}
	h := speakerHandler{SpeakerStorage: SpeakerStorage}
	m.HandleFunc("/{ID}/", h.GetSpeaker).Methods("GET")
	m.HandleFunc("/", h.AddSpeaker).Methods("POST")
	m.HandleFunc("/list", h.ListSpeakers).Methods("GET")
	m.HandleFunc("/{ID}/update", h.UpdateSpeaker).Methods("POST")
	m.HandleFunc("/{ID}/delete", h.DeleteSpeaker).Methods("GET")
	m.HandleFunc("/form/update", updateSpeakerForm).Methods("GET")

	return nil
}

type speakerHandler struct {
	SpeakerStorage SpeakerStore
}

func (h *speakerHandler) GetSpeaker(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	vars := mux.Vars(r)
	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "ID not valid: %v", vars["ID"])
		return
	}

	speaker, err := h.SpeakerStorage.GetSpeaker(ctx, ID)
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

func (h *speakerHandler) AddSpeaker(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, fmt.Sprint("/public/#/add_speaker"))
		fmt.Fprintf(w, url)
		return
	}

	speaker := Speaker{}
	err := json.NewDecoder(r.Body).Decode(&speaker)
	if err != nil {
		log.Errorf(ctx, "Couldn't decode JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if speaker.Name == "" || speaker.Surname == "" || speaker.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Fields Name, Surname and Email are mandatory!")
		return
	}

	speaker.Owner = u.Email

	id, err := h.SpeakerStorage.AddSpeaker(ctx, &speaker)
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", id)
}

func (h *speakerHandler) UpdateSpeaker(w http.ResponseWriter, r *http.Request) {
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
		//Make you sure that ulr is correct.
		url, _ := user.LoginURL(ctx, fmt.Sprint("/public/#/update_speaker"))
		fmt.Fprint(w, url)
		return
	}

	suf := SpeakerForm{}
	err = json.NewDecoder(r.Body).Decode(&suf)
	log.Debugf(ctx, "Body: %s", r.Body)
	if err != nil {
		log.Errorf(ctx, "Couldn't decode JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	speaker, err := h.SpeakerStorage.GetSpeaker(ctx, ID)
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
	if speaker.Owner != u.Email && !u.Admin {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "You're not the owner nor the admin.")
		return
	}
	//TODO: Update speaker with function.
	if suf.Name != "" {
		speaker.Name = suf.Name
	}

	if suf.Surname != "" {
		speaker.Surname = suf.Surname
	}

	if suf.Email != "" {
		speaker.Email = suf.Email
	}

	if suf.Company != "" {
		speaker.Company = suf.Company
	}

	if suf.About != "" {
		speaker.About = suf.About
	}

	err = h.SpeakerStorage.PutSpeaker(ctx, ID, &speaker)
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

func (h *speakerHandler) DeleteSpeaker(w http.ResponseWriter, r *http.Request) {
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
		url, _ := user.LoginURL(ctx, fmt.Sprintf("/public/#/delete_speaker/%v/", ID))
		fmt.Fprintf(w, url)
		return
	}

	speaker, err := h.SpeakerStorage.GetSpeaker(ctx, ID)
	if err == datastore.ErrNoSuchEntity {
		fmt.Fprint(w, "Speaker not found.")
		return
	}
	if err != nil {
		log.Errorf(ctx, "Can't get speaker: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if speaker.Owner != u.Email && !u.Admin {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "You're not the owner nor the admin.")
		return
	}

	err = h.SpeakerStorage.DeleteSpeaker(ctx, ID)
	if err != nil {
		log.Errorf(ctx, "Can't delete speaker: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusTeapot)
	fmt.Fprint(w, "Speaker deleted successfully.")
}

func (h *speakerHandler) ListSpeakers(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	IDs, speakers, err := h.SpeakerStorage.GetAllSpeakers(ctx)
	if err != nil {
		log.Errorf(ctx, "Can't get speakers: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	speakersPublicView := make([]SpeakerPublicView, 0, len(speakers))
	for index, speaker := range speakers {
		speakersPublicView = append(speakersPublicView, speaker.GetPublicView(IDs[index]))
	}

	err = WriteSpeakersPublicView(speakersPublicView, w)
	if err != nil {
		log.Errorf(ctx, "Failed to write speakers slice: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
