package MeetupRest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	Title       string
	Description string
	Speakers    string
	Voters      []string
}

type PresentationForm struct {
	Title       string
	Description string
	Speakers    []int64
}

type PresentationPublicView struct {
	Key         int64
	Title       string
	Description string
	Speakers    []string
	Votes       int
}

type PresentationStore interface {
	SpeakerStore
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
func RegisterPresentationRoutes(m *mux.Router, Storage PresentationStore) error {
	if m == nil {
		return errors.New("m may not be nil when registering presentation routes")
	}
	h := presentationHandler{Storage: Storage}
	m.HandleFunc("/{ID}/", h.getPresentation).Methods("GET")
	m.HandleFunc("/", h.addPresentation).Methods("POST")
	m.HandleFunc("/", h.removePresentation).Methods("DELETE")
	m.HandleFunc("/{ID}/update", h.updatePresentation).Methods("POST")
	m.HandleFunc("/list", h.listPresentations).Methods("GET")
	m.HandleFunc("/{ID}/upvote", h.upvotePresentation).Methods("GET")
	m.HandleFunc("/{ID}/downvote", h.downvotePresentation).Methods("GET")
	m.HandleFunc("/{ID}/hasUpvoted", h.hasUpvoted).Methods("GET")

	return nil
}

type presentationHandler struct {
	Storage PresentationStore
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

	presentation, err := h.Storage.GetPresentation(ctx, ID)
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

	//Decode Speakers
	speakersId := stringToIntSlice(presentation.Speakers)
	speakers := []string{}
	for i := range speakersId {
		speaker, err := h.Storage.GetSpeaker(ctx, speakersId[i])
		if err != nil {
			log.Infof(ctx, "Couldn't get speaker with key: %v, error: %v", speakersId[i], err)
			continue
		}
		speakers = append(speakers, speaker.GetSpeakerFullName())
	}

	presentationPublicView := presentation.GetPublicView(ID, speakers)
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

	puf := PresentationForm{}
	err := json.NewDecoder(r.Body).Decode(&puf)
	log.Debugf(ctx, "Body: %s", r.Body)
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

	p := Presentation{}
	p.Title = puf.Title
	p.Description = puf.Description
	p.Speakers = sliceIntToString(puf.Speakers)

	ID, err := h.Storage.AddPresentation(ctx, &p)
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", ID)
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

	puf := PresentationForm{}
	err = json.NewDecoder(r.Body).Decode(&puf)
	log.Debugf(ctx, "Body: %s", r.Body)
	if err != nil {
		log.Errorf(ctx, "Couldn't decode JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	presentation, err := h.Storage.GetPresentation(ctx, ID)
	if err == datastore.ErrNoSuchEntity {
		fmt.Fprint(w, "No presentation with ID: %v", ID)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(ctx, "Couldn't get presentation with key: %v, error: %v", ID, err)
		return
	}

	if puf.Title != presentation.Title {
		presentation.Title = puf.Title
	}

	if puf.Description != presentation.Description {
		presentation.Description = puf.Description
	}

	// This condidtion is correct?
	if puf.Speakers[0] != 0 {
		presentation.Speakers = sliceIntToString(puf.Speakers)
	}

	err = h.Storage.PutPresentation(ctx, ID, &presentation)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(ctx, "Couldn't put presentation into datastore: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "Presentation Updated!")
}

func (h *presentationHandler) removePresentation(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Errorf(ctx, "Can't parse query: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	presentationID, ok := params["PresentationId"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Presentation ID must be provided.")
		return
	}

	keyInt, err := strconv.ParseInt(presentationID[0], 10, 32)

	err = h.Storage.DeletePresentation(ctx, keyInt)
	if err != nil {
		log.Errorf(ctx, "Can't delete meetup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusTeapot)
	fmt.Fprintf(w, "Presentation deleted successfully. %s", keyInt)
}

func (h *presentationHandler) listPresentations(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	IDs, presentations, err := h.Storage.GetAllPresentations(ctx)
	if err != nil {
		log.Errorf(ctx, "Can't get presentations: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	presentationsPublicView := make([]PresentationPublicView, 0, len(presentations))

	for idx, presentation := range presentations {
		// Fill presentation public views ;)
		speakersId := stringToIntSlice(presentation.Speakers)
		speakers := []string{}
		for i := range speakersId {
			speaker, err := h.Storage.GetSpeaker(ctx, speakersId[i])
			if err != nil {
				log.Infof(ctx, "Couldn't get speaker with key: %v, error: %v", speakersId[i], err)
				continue
			}
			speakers = append(speakers, speaker.GetSpeakerFullName())
			presentationsPublicView = append(presentationsPublicView, presentation.GetPublicView(IDs[idx], speakers))
		}

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

	presentation, err := h.Storage.GetPresentation(ctx, ID)
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

	err = h.Storage.PutPresentation(ctx, ID, &presentation)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(ctx, "Couldn't put presentation into datastore: %v", err)
	}
	fmt.Fprint(w, "Upvoted!")
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

	presentation, err := h.Storage.GetPresentation(ctx, ID)
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

	err = h.Storage.PutPresentation(ctx, ID, &presentation)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(ctx, "Couldn't put presentation into datastore: %v", err)
	}
	fmt.Fprint(w, "Undone upvote!")
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

	presentation, err := h.Storage.GetPresentation(ctx, ID)
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

func (p *Presentation) GetPublicView(key int64, speakers []string) PresentationPublicView {
	return PresentationPublicView{
		Key:         key,
		Title:       p.Title,
		Description: p.Description,
		Speakers:    speakers,
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

func sliceIntToString(ints []int64) string {
	tmpStringArray := []string{}
	for i := range ints {
		numString := strconv.FormatInt(ints[i], 10)
		tmpStringArray = append(tmpStringArray, numString)
	}
	return strings.Join(tmpStringArray, ",")
}

func stringToIntSlice(inputString string) []int64 {
	intStrings := strings.Split(inputString, ",")
	ints := []int64{}
	for i := range intStrings {
		item := intStrings[i]
		itemInt, err := strconv.ParseInt(item, 10, 64)
		if err != nil {
			continue
		}
		ints = append(ints, itemInt)
	}
	return ints
}
