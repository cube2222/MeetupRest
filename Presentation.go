package MeetupRest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type PresentationPublicView struct {
	Key         int64
	Title       string
	Description string
	Speaker     string
	Votes       int
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
func RegisterPresentationRoutes(m *mux.Router) error {
	if m == nil {
		return errors.New("m may not be nil when registering presentation routes")
	}
	m.HandleFunc("/", getPresentation).Methods("GET")
	m.HandleFunc("/", addPresentation).Methods("POST")
	m.HandleFunc("/", removePresentation).Methods("DELETE")
	m.HandleFunc("/list", listPresentations).Methods("GET")
	m.HandleFunc("/form/add", addPresentationForm).Methods("GET")
	m.HandleFunc("/form/delete", removePresentationForm).Methods("GET")
	m.HandleFunc("/{ID}/upvote", upvotePresentation).Methods("GET")
	m.HandleFunc("/{ID}/downvote", downvotePresentation).Methods("GET")

	return nil
}

func getPresentation(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Errorf(ctx, "Can't parse query: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	q := datastore.NewQuery(datastorePresentationsKind).Limit(1)

	if title, ok := params["title"]; ok {
		q = q.Filter("Title=", title[0])
	}

	if speaker, ok := params["speaker"]; ok {
		q = q.Filter("Speaker=", speaker[0])
	}

	if description, ok := params["description"]; ok {
		q = q.Filter("Description=", description[0])
	}

	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	t := q.Run(newCtx)
	done()
	myPresentation := Presentation{}
	key, err := t.Next(&myPresentation)
	// No speaker retrieved
	if err == datastore.Done {
		fmt.Fprint(w, "No Presentation found.")
		return
	}
	// Some other error
	if err != nil {
		log.Errorf(ctx, "Can't get presentation: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newCtx, done = context.WithTimeout(ctx, time.Second*2)
	speakerRetrieved, err := GetSpeakerByKey(newCtx, myPresentation.Speaker)
	done()

	if err != nil {
		log.Infof(ctx, "Couldn't get speaker with key: %v, error: %v", key.IntID(), err)
	}
	speakerPublicView := myPresentation.GetPublicView(key.IntID(), speakerRetrieved.GetSpeakerFullName())
	err = speakerPublicView.WriteTo(w)
	if err != nil {
		log.Errorf(ctx, "Failed to write presentation: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func addPresentation(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	err := r.ParseForm()
	if err != nil {
		log.Errorf(ctx, "Couldn't parse form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p := &Presentation{}

	decoder := schema.NewDecoder()
	decoder.Decode(p, r.PostForm)

	if p.Title == "" || p.Speaker == 0 || p.Description == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Fields Title, Speaker and Description are mandatory!")
		return
	}

	key := datastore.NewKey(ctx, datastorePresentationsKind, "", 0, nil)
	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	id, err := datastore.Put(newCtx, key, p)
	done()
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", id.IntID())
}

func removePresentation(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

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

	key := datastore.NewKey(ctx, datastorePresentationsKind, "", keyInt, nil)
	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	err = datastore.Delete(newCtx, key)
	done()
	if err != nil {
		log.Errorf(ctx, "Can't delete meetup: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusTeapot)
	fmt.Fprintf(w, "Presentation deleted successfully. %s", key.AppID())
}

func addPresentationForm(w http.ResponseWriter, r *http.Request) {
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

func listPresentations(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Errorf(ctx, "Can't parse query: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	q := datastore.NewQuery(datastorePresentationsKind)

	speaker, okSpeaker := params["speaker"]

	if okSpeaker {
		speakerID, err := strconv.ParseInt(speaker[0], 10, 64)
		if err != nil {
			log.Errorf(ctx, "Can't parse to int64, error: %v", err)
			return
		}
		q = q.Filter("Speaker=", speakerID)
	}

	presentations := make([]Presentation, 0, 10)
	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	keys, err := q.GetAll(newCtx, &presentations)
	done()
	if err != nil {
		log.Errorf(ctx, "Can't get presentations: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	speakers := make([]Speaker, len(presentations))
	for i := 0; i < len(presentations); i++ {
		newCtx, done = context.WithTimeout(ctx, time.Second*2)
		speakers[i], err = GetSpeakerByKey(newCtx, presentations[i].Speaker)
		done()
		if err != nil {
			log.Infof(ctx, "Couldn't get speaker with key: %v, error: %v", presentations[i].Speaker, err)
		}
	}

	// We don't want people to see voters. Just the vote count.
	presentationsPublicView := make([]PresentationPublicView, 0, len(presentations))
	for index, presentation := range presentations {
		presentationsPublicView = append(presentationsPublicView, presentation.GetPublicView(
			keys[index].IntID(),
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

func upvotePresentation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	ctx := appengine.NewContext(r)
	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, fmt.Sprintf("/presentation/%v/upvote", vars["ID"]))
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
	}

	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Please provide a valid ID.")
		return
	}

	key := datastore.NewKey(ctx, datastorePresentationsKind, "", ID, nil)

	presentation := Presentation{}

	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	err = datastore.Get(newCtx, key, &presentation)
	done()
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

	newCtx, done = context.WithTimeout(ctx, time.Second*2)
	_, err = datastore.Put(newCtx, key, &presentation)
	done()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(ctx, "Couldn't put presentation into datastore: %v", err)
	}
	fmt.Fprint(w, "Upvoted!")
}

func downvotePresentation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	ctx := appengine.NewContext(r)
	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, fmt.Sprintf("/presentation/%v/downvote", vars["ID"]))
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
	}

	ID, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Please provide a valid ID.")
		return
	}

	key := datastore.NewKey(ctx, datastorePresentationsKind, "", ID, nil)

	presentation := Presentation{}

	newCtx, done := context.WithTimeout(ctx, time.Second*2)
	err = datastore.Get(newCtx, key, &presentation)
	done()
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

	newCtx, done = context.WithTimeout(ctx, time.Second*2)
	_, err = datastore.Put(newCtx, key, &presentation)
	done()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(ctx, "Couldn't put presentation into datastore: %v", err)
	}
	fmt.Fprint(w, "Upvoted!")
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
