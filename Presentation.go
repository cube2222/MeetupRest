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

const datastorePresentationskind = "Presentations"

type Presentation struct {
	Title       string
	Description string
	Speaker     string
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

// Get the handler which contains all the presentation handling routes and the corresponding handlers.
func RegisterPresentationRoutes(m *mux.Router) error {
	if m == nil {
		errors.New("m may not be nil when registering presentation routes")
	}
	m.HandleFunc("/", getPresentation).Methods("GET")
	m.HandleFunc("/", addPresentation).Methods("POST")
	m.HandleFunc("/list", listPresentations).Methods("GET")
	m.HandleFunc("/form/add", addPresentationForm).Methods("GET")

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

	q := datastore.NewQuery(datastorePresentationskind).Limit(1)

	if title, ok := params["title"]; ok == true {
		q = q.Filter("Title=", title[0])
	}

	if speaker, ok := params["speaker"]; ok == true {
		q = q.Filter("Speaker=", speaker[0])
	}

	if description, ok := params["description"]; ok == true {
		q = q.Filter("Description=", description[0])
	}

	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	t := q.Run(newCtx)
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
	data, err := json.Marshal(&PresentationPublicView{
		Key:         key.IntID(),
		Title:       myPresentation.Title,
		Description: myPresentation.Description,
		Speaker:     myPresentation.Speaker,
		Votes:       len(myPresentation.Voters),
	})
	if err != nil {
		log.Errorf(ctx, "Failed to serialize presentation: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	io.Copy(w, bytes.NewReader(data))
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

	if p.Title == "" || p.Speaker == "" || p.Description == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Fields Title, Speaker and Description are mandatory!")
		return
	}

	key := datastore.NewKey(ctx, datastorePresentationskind, "", 0, nil)
	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	id, err := datastore.Put(newCtx, key, p)
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", id.IntID())
}

func addPresentationForm(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	speakers := make([]Speaker, 0, 10)

	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	keys, err := datastore.NewQuery(datastoreSpeakersKind).GetAll(newCtx, &speakers)
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

	var placesPageTmpl *template.Template = template.Must(template.New("PlacesPage").Parse(HTML))

	buf := bytes.Buffer{}
	if err := placesPageTmpl.Execute(&buf, options); err != nil {
		fmt.Println("Failed to build page", err)
	} else {
		fmt.Fprint(w, buf.String())
	}
}

const HTML = `
	<h1>Adding Presentation Form</h1>
        <form action="/presentation/" method="POST">
		Title: <input type="text" name="Title"><br>
		Description: <textarea name="Description"></textarea><br>"+
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

// func removePresentationForm(w http.ResponseWriter, r *http.ReadRequest) {
// 	ctx := appengine.NewContext(r)

// 	presentations := make([]Presentation, 0, 10)
// 	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
// 	keys, err := datastore.NewQuery(datastorePresentationskind).GetAll(newCtx, &presentations)
// 	if err != nil {
// 		log.Errorf(ctx, "Can't get Presentations: %v", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}

// }

func listPresentations(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Errorf(ctx, "Can't parse query: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	q := datastore.NewQuery(datastorePresentationskind)

	speaker, okSpeaker := params["speaker"]

	if okSpeaker {
		q = q.Filter("Speaker=", speaker[0])
	}

	presentations := make([]Presentation, 0, 10)
	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	keys, err := q.GetAll(newCtx, &presentations)
	if err != nil {
		log.Errorf(ctx, "Can't get presentations: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// We don't want people to see voters. Just the vote count.
	presentationsPublicView := make([]PresentationPublicView, 0, len(presentations))
	for index, presentation := range presentations {
		presentationsPublicView = append(presentationsPublicView, PresentationPublicView{
			Key:         keys[index].IntID(),
			Title:       presentation.Title,
			Description: presentation.Description,
			Speaker:     presentation.Speaker,
			Votes:       len(presentation.Voters),
		})
	}

	data, err := json.Marshal(&presentationsPublicView)
	if err != nil {
		log.Errorf(ctx, "Failed to serialize presentations slice: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	io.Copy(w, bytes.NewReader(data))
}

func upvotePresentation(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, "/")
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
	}
	url, _ := user.LogoutURL(ctx, "/")
	fmt.Fprintf(w, `Welcome, %s! (<a href="%s">sign out</a>)`, u, url)
}
