package MeetupRest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

const datastorePresentationsKind = "Presentations"

type Presentation struct {
	Owner       string
	Title       string
	Description string
	Speakers    []string
	Voters      []string
}

type PresentationForm struct {
	Title       string
	Description string
	Speakers    string
}

type PresentationPublicView struct {
	Key         int64
	Title       string
	Description string
	Speakers    []string
	Votes       int
}

type PresentationStore interface {
	GetPresentation(ctx context.Context, id int64) (Presentation, error)
	GetAllPresentations(ctx context.Context) ([]int64, []Presentation, error)
	PutPresentation(ctx context.Context, id int64, presentation *Presentation) error
	AddPresentation(ctx context.Context, presentation *Presentation) (int64, error)
	DeletePresentation(ctx context.Context, id int64) error
}

type Option struct {
	Value, Text string
}

// Get the handler which contains all the presentation handling routes and the corresponding handlers.
func RegisterPresentationRoutes(m *mux.Router, PresentationStorage PresentationStore, SpeakerStorage SpeakerStore, MeetupAPIUpdateFunction func() error) error {
	if m == nil {
		return errors.New("m may not be nil when registering presentation routes")
	}
	h := presentationHandler{PresentationStorage: PresentationStorage, SpeakerStorage: SpeakerStorage, MeetupAPIUpdateFunction: MeetupAPIUpdateFunction}
	m.HandleFunc("/{ID}/", h.getPresentation).Methods("GET")
	m.HandleFunc("/", h.addPresentation).Methods("POST")
	m.HandleFunc("/{ID}/delete", h.deletePresentation).Methods("GET")
	m.HandleFunc("/{ID}/update", h.updatePresentation).Methods("POST")
	m.HandleFunc("/list", h.listPresentations).Methods("GET")
	m.HandleFunc("/{ID}/upvote", h.upvotePresentation).Methods("GET")
	m.HandleFunc("/{ID}/downvote", h.downvotePresentation).Methods("GET")
	m.HandleFunc("/{ID}/hasUpvoted", h.hasUpvoted).Methods("GET")

	return nil
}

type presentationHandler struct {
	PresentationStorage     PresentationStore
	SpeakerStorage          SpeakerStore
	MeetupAPIUpdateFunction func() error
}

func (h *presentationHandler) getPresentation(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	vars := mux.Vars(r)
	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "ID not valid: %v", vars["ID"])
	}

	presentation, err := h.PresentationStorage.GetPresentation(ctx, ID)
	if err == datastore.ErrNoSuchEntity {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Couldn't find presentation with id: %v", ID)
		return
	}
	if err != nil {
		log.Errorf(ctx, "Couldn't get presentation: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	presentationPublicView := presentation.GetPublicView(ID)
	err = presentationPublicView.WriteTo(w)
	if err != nil {
		log.Errorf(ctx, "Failed to write presentation: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *presentationHandler) addPresentation(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, fmt.Sprint("/presentation/form/add"))
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
	}

	puf := PresentationForm{}
	err := json.NewDecoder(r.Body).Decode(&puf)
	if err != nil {
		log.Errorf(ctx, "Couldn't decode JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if puf.Title == "" || puf.Speakers[0] == 0 || puf.Description == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Fields Title, Speakers and Description are mandatory!")
		return
	}

	presentation := Presentation{}
	presentation.Title = puf.Title
	presentation.Description = puf.Description
	presentation.Speakers = strings.Split(puf.Speakers, ",")
	for index, item := range presentation.Speakers {
		presentation.Speakers[index] = strings.Trim(item, " ")
	}
	presentation.Owner = u.Email

	ID, err := h.PresentationStorage.AddPresentation(ctx, &presentation)
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", ID)

	err = h.MeetupAPIUpdateFunction()
	if err != nil {
		log.Errorf(ctx, "Error when updating meetup API: %v", err)
		return
	}
}

func (h *presentationHandler) updatePresentation(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	vars := mux.Vars(r)
	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Please provide a valid ID.")
		return
	}

	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, fmt.Sprint("/presentation/form/update"))
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
	}

	puf := PresentationForm{}
	err = json.NewDecoder(r.Body).Decode(&puf)
	log.Debugf(ctx, "Body: %s", r.Body)
	if err != nil {
		log.Errorf(ctx, "Couldn't decode JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	presentation, err := h.PresentationStorage.GetPresentation(ctx, ID)
	if err == datastore.ErrNoSuchEntity {
		fmt.Fprint(w, "No presentation with ID: %v", ID)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(ctx, "Couldn't get presentation with key: %v, error: %v", ID, err)
		return
	}

	// Check if it's the owner
	if presentation.Owner != u.Email && !u.Admin {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "You're not the owner nor the admin.")
		return
	}

	if puf.Title != presentation.Title {
		presentation.Title = puf.Title
	}

	if puf.Description != presentation.Description {
		presentation.Description = puf.Description
	}

	speakers := strings.Split(puf.Speakers, ",")
	same := true
	for _, item := range speakers {
		found := false
		for _, item2 := range presentation.Speakers {
			if item == item2 {
				found = true
				break
			}
		}
		if !found {
			same = false
			break
		}
	}

	if !same {
		presentation.Speakers = speakers
	}

	err = h.PresentationStorage.PutPresentation(ctx, ID, &presentation)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(ctx, "Couldn't put presentation into datastore: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "Presentation Updated!")

	err = h.MeetupAPIUpdateFunction()
	if err != nil {
		log.Errorf(ctx, "Error when updating meetup API: %v", err)
		return
	}
}

func (h *presentationHandler) deletePresentation(w http.ResponseWriter, r *http.Request) {
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
		url, _ := user.LoginURL(ctx, fmt.Sprintf("/presentation/%v/delete", ID))
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
	}

	presentation, err := h.PresentationStorage.GetPresentation(ctx, ID)
	if err == datastore.ErrNoSuchEntity {
		fmt.Fprint(w, "Speaker not found.")
		return
	}
	if err != nil {
		log.Errorf(ctx, "Can't get speaker: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if presentation.Owner != u.Email && !u.Admin {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "You're not the owner nor the admin.")
		return
	}

	err = h.PresentationStorage.DeletePresentation(ctx, ID)
	if err != nil {
		log.Errorf(ctx, "Can't delete meetup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusTeapot)
	fmt.Fprintf(w, "Presentation deleted successfully. %v", ID)

	err = h.MeetupAPIUpdateFunction()
	if err != nil {
		log.Errorf(ctx, "Error when updating meetup API: %v", err)
		return
	}
}

func (h *presentationHandler) listPresentations(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	IDs, presentations, err := h.PresentationStorage.GetAllPresentations(ctx)
	if err != nil {
		log.Errorf(ctx, "Can't get presentations: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	presentationsPublicView := make([]PresentationPublicView, 0, len(presentations))

	for idx, presentation := range presentations {
		presentationsPublicView = append(presentationsPublicView, presentation.GetPublicView(IDs[idx]))
	}

	err = WritePresentationsPublicView(presentationsPublicView, w)
	if err != nil {
		log.Errorf(ctx, "Failed to write presentations slice: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *presentationHandler) upvotePresentation(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	vars := mux.Vars(r)
	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Please provide a valid ID.")
		return
	}

	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, fmt.Sprintf("/presentation/%v/upvote", vars["ID"]))
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
	}

	presentation, err := h.PresentationStorage.GetPresentation(ctx, ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(ctx, "Couldn't get presentation with key: %v, error: %v", ID, err)
		return
	}

	if contains(presentation.Voters, u.Email) {
		fmt.Fprint(w, "Sorry, you already upvoted this presentation.")
		return
	}

	presentation.Voters = append(presentation.Voters, u.Email)

	err = h.PresentationStorage.PutPresentation(ctx, ID, &presentation)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(ctx, "Couldn't put presentation into datastore: %v", err)
	}
	fmt.Fprint(w, "Upvoted!")

	err = h.MeetupAPIUpdateFunction()
	if err != nil {
		log.Errorf(ctx, "Error when updating meetup API: %v", err)
		return
	}
}

func (h *presentationHandler) downvotePresentation(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	vars := mux.Vars(r)
	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Please provide a valid ID.")
		return
	}

	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, fmt.Sprintf("/presentation/%v/downvote", vars["ID"]))
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
	}

	presentation, err := h.PresentationStorage.GetPresentation(ctx, ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(ctx, "Couldn't get presentation with key: %v, error: %v", ID, err)
		return
	}

	if !contains(presentation.Voters, u.Email) {
		fmt.Fprint(w, "Sorry, you haven't upvoted this presentation.")
		return
	}

	for i := 0; i < len(presentation.Voters); i++ {
		if presentation.Voters[i] == u.Email {
			presentation.Voters = append(presentation.Voters[:i], presentation.Voters[i+1:]...)
			break
		}
	}

	err = h.PresentationStorage.PutPresentation(ctx, ID, &presentation)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(ctx, "Couldn't put presentation into datastore: %v", err)
	}
	fmt.Fprint(w, "Undone upvote!")

	err = h.MeetupAPIUpdateFunction()
	if err != nil {
		log.Errorf(ctx, "Error when updating meetup API: %v", err)
		return
	}
}

func (h *presentationHandler) hasUpvoted(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	u := user.Current(ctx)
	if u == nil {
		fmt.Fprint(w, "false")
		return
	}

	vars := mux.Vars(r)
	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Please provide a valid ID.")
		return
	}

	presentation, err := h.PresentationStorage.GetPresentation(ctx, ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(ctx, "Couldn't get presentation with key: %v, error: %v", ID, err)
		return
	}

	if contains(presentation.Voters, u.Email) {
		fmt.Fprint(w, "true")
		return
	} else {
		fmt.Fprint(w, "false")
		return
	}
}

func contains(slice []string, text string) bool {
	for _, item := range slice {
		if item == text {
			return true
		}
	}
	return false
}

func (p *Presentation) GetPublicView(key int64) PresentationPublicView {
	return PresentationPublicView{
		Key:         key,
		Title:       p.Title,
		Description: p.Description,
		Speakers:    p.Speakers,
		Votes:       len(p.Voters),
	}
}

func (p *Presentation) WriteTo(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

func (p *PresentationPublicView) WriteTo(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

func WritePresentationsPublicView(presentations []PresentationPublicView, w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(presentations)
}
