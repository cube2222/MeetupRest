package MeetupRest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"io"
	"net/http"
	"net/url"
)

const kindPresentations = "Presentations"

type Presentation struct {
	Title       string
	Description string
	Speaker     string
	VoteCount   int
}

func GetPresentationHandler() http.Handler {
	m := mux.NewRouter()
	m.HandleFunc("/presentation", getSpeaker).Methods("GET")
	m.HandleFunc("/presentation", addSpeaker).Methods("POST")

	return m
}

func getPresentation(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Errorf(ctx, "Can't parse query: %v", err)
		// TODO: http code
		fmt.Fprintf(w, "Can't parse query: %v", err)
		return
	}

	q := datastore.NewQuery(kindPresentations).Limit(1)

	if title, ok := params["title"]; ok == true {
		q = q.Filter("Title=", title[0])
	}

	if speaker, ok := params["speaker"]; ok == true {
		q = q.Filter("Speaker=", speaker[0])
	}

	if description, ok := params["description"]; ok == true {
		q = q.Filter("Description=", description[0])
	}

	t := q.Run(ctx)
	myPresentation := Presentation{}

	_, err = t.Next(&myPresentation)
	if err == datastore.Done {
		fmt.Fprint(w, "No Presentation found.")
		return
	}
	data, err := json.Marshal(&myPresentation)
	if err != nil {
		log.Errorf(ctx, "Failed to serialize presentation: %v", err)
		// TODO: http code
		fmt.Fprintf(w, "Failed to serialize presentation: %v", err)
		return
	}
	io.Copy(w, bytes.NewBuffer(data))
}

func addPresentation(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	err := r.ParseForm()
	if err != nil {
		log.Errorf(ctx, "Couldn't parse form: Presentation")
		// TODO: Status Code
		fmt.Fprint(w, "Couldn't parse form: Presentation")
		return
	}

	p := &Presentation{}

	decoder := schema.NewDecoder()
	decoder.Decode(p, r.PostForm)

	if p.Title == "" || p.Speaker == "" || p.Description == "" {
		fmt.Fprint(w, "Fields Title, Speaker and Description are mandatory!")
	}

	key := datastore.NewKey(ctx, kindPresentations, "", 0, nil)
	//newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	id, err := datastore.Put(ctx, key, p)
	if err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", id.IntID())
}
