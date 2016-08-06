package MeetupRest

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"google.golang.org/cloud/datastore"
	"net/http"
)

func init() {
	ctx := context.Background()

	m := mux.NewRouter()
	m.HandleFunc("/", handler)
	http.Handle("/", m)

	dsClient, err := datastore.NewClient(ctx, "meetuprest")
	if err != nil {

	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world from gorilla!!!")
}
