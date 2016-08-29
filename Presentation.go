package MeetupRest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

const datastorePresentationsKind = "Presentations"

type Presentation struct {
	Title       string
	Description string
	Speaker     int64
	Voters      []string
}

type PresentationForm struct {
	Title       string
	Description string
}

type PresentationPublicView struct {
	Key         int64
	Title       string
	Description string
	Speaker     string
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

const HTMLAddForm = `
	<h1>Adding Presentation Form</h1>
        <form action="/presentation/" method="POST">
		Title: <input type="text" name="Title"><br>
		Description: <textarea name="Description"></textarea><br>
            <div>
                <label>Speaker:</label>
                <select name="Speaker">
                    {{range .}}
                    <option value="{{.Value}}">{{.Text}}</option>
                    {{end}}
                </select>
            </div>
            <input type="submit" value="Save">
        </form>
`

const HTMLDeleteForm = `
	<h1>Deleting Presentation Form</h1>
        <form action="/presentation/" method="DELETE">
            <div>
                <label>By Title:</label>
                <select name="PresentationId">
                    {{range .}}
                    <option value="{{.Value}}">{{.Text}}</option>
                    {{end}}
                </select>
            </div>
            <input type="submit" value="Remove">
        </form>
`

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
	m.HandleFunc("/form/add", addPresentationForm).Methods("GET")
	m.HandleFunc("/form/{ID}/delete", removePresentationForm).Methods("GET")
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

	speaker, err := h.Storage.GetSpeaker(ctx, presentation.Speaker)
	if err != nil {
		log.Infof(ctx, "Couldn't get speaker with key: %v, error: %v", ID, err)
	}

	presentationPublicView := presentation.GetPublicView(ID, speaker.GetSpeakerFullName())
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

	err := r.ParseForm()
	if err != nil {
		log.Errorf(ctx, "Couldn't parse form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p := Presentation{}

	decoder := schema.NewDecoder()
	decoder.Decode(&p, r.PostForm)

	if p.Title == "" || p.Speaker == 0 || p.Description == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Fields Title, Speaker and Description are mandatory!")
		return
	}

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

	if puf.Title != "" {
		presentation.Title = puf.Title
	}

	if puf.Description != "" {
		presentation.Description = puf.Description
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

func addPresentationForm(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	speakers := make([]Speaker, 0, 10)

	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	keys, err := datastore.NewQuery(datastoreSpeakersKind).GetAll(newCtx, &speakers)
	done()
	if err != nil {
		log.Errorf(ctx, "Can't get speakers: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	options := make([]Option, 0, len(speakers))
	for index, speaker := range speakers {
		options = append(options, Option{
			Value: strconv.FormatInt(keys[index].IntID(), 10),
			Text:  speaker.Name + " " + speaker.Surname,
		})
	}

	placesPageTmpl := template.Must(template.New("PlacesPage").Parse(HTMLAddForm))

	buf := bytes.Buffer{}
	if err := placesPageTmpl.Execute(&buf, options); err != nil {
		fmt.Println("Failed to build page", err)
	} else {
		fmt.Fprint(w, buf.String())
	}
}

func removePresentationForm(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	presentations := make([]Presentation, 0, 10)
	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	keys, err := datastore.NewQuery(datastorePresentationsKind).GetAll(newCtx, &presentations)
	done()
	if err != nil {
		log.Errorf(ctx, "Can't get Presentations: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	options := make([]Option, 0, len(presentations))
	for index, presentation := range presentations {
		options = append(options, Option{
			Value: strconv.FormatInt(keys[index].IntID(), 10),
			Text:  presentation.Title,
		})
	}

	placesPageTmpl := template.Must(template.New("PlacesPage").Parse(HTMLDeleteForm))

	buf := bytes.Buffer{}
	if err := placesPageTmpl.Execute(&buf, options); err != nil {
		fmt.Println("Failed to build page", err)
	} else {
		fmt.Fprint(w, buf.String())
	}
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

	speakers := make([]Speaker, len(presentations))
	for i := 0; i < len(presentations); i++ {
		speakers[i], err = h.Storage.GetSpeaker(ctx, presentations[i].Speaker)
		if err != nil {
			log.Errorf(ctx, "Couldn't get speaker with key: %v, error: %v", presentations[i].Speaker, err)
		}
	}

	// We don't want people to see voters. Just the vote count.
	presentationsPublicView := make([]PresentationPublicView, 0, len(presentations))
	for index, presentation := range presentations {
		presentationsPublicView = append(presentationsPublicView, presentation.GetPublicView(
			IDs[index],
			speakers[index].GetSpeakerFullName(),
		))
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

func (p *Presentation) GetPublicView(key int64, speaker string) PresentationPublicView {
	return PresentationPublicView{
		Key:         key,
		Title:       p.Title,
		Description: p.Description,
		Speaker:     speaker,
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
