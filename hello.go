package MeetupRest

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/cloud/datastore"
	"net/http"
	"strconv"
	"time"
)

type Human struct {
	Name string
	Age  int
}

func init() {
	ctx := context.Background()

	dsClient, err := datastore.NewClient(ctx, "meetuprest")
	if err != nil {
		log.Errorf(ctx, "Could not create Datastore Client:", err)
		m.ha
	}

	m := mux.NewRouter()
	m.Handle("/{name}/{age}", dsHandler{dsClient, ctx})
	http.Handle("/", m)
}

type dsHandler struct {
	dsClient *datastore.Client
	ctx      context.Context
}

func (h dsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	aectx := appengine.NewContext(r)
	variables := mux.Vars(r)
	e := &Human{}

	e.Name = variables["name"]
	age, _ := strconv.Atoi(variables["age"])
	e.Age = age

	k := datastore.NewKey(h.ctx, "People", "", 0, nil)
	newCtx, _ := context.WithTimeout(h.ctx, time.Second*2)
	id, err := h.dsClient.Put(newCtx, k, e)
	if err != nil {
		log.Errorf(aectx, "Can't create datastore object: %v", err)
		return
	}

	fmt.Fprintf(w, "Your key: %v", id.ID())
}
