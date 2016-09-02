package MeetupRest

import (
	"net/http"

	"fmt"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/user"
	"net/url"
	"time"
)

var defaultRequestTimeout = time.Second * 4

func init() {

	m := mux.NewRouter()
	Storage := GoogleDatastoreStore{}

	MeetupAPIUpdateFunction := getMeetupUpdateFunction(&Storage)

	s := m.PathPrefix("/speaker").Subrouter()
	err := RegisterSpeakerRoutes(s, &Storage)

	s = m.PathPrefix("/presentation").Subrouter()
	err = RegisterPresentationRoutes(s, &Storage, MeetupAPIUpdateFunction)

	s = m.PathPrefix("/meetup").Subrouter()
	err = RegisterMeetupRoutes(s, &Storage, MeetupAPIUpdateFunction)

	s = m.PathPrefix("/metadata").Subrouter()
	err = RegisterMetadataRoutes(s, &Storage)

	m.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("public/"))))

	m.HandleFunc("/isLoggedIn", isLoggedIn)
	m.HandleFunc("/getLoginAddress", getLoginAddress)

	if err != nil {
		panic(err)
	}

	http.Handle("/", m)
}

func isLoggedIn(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	u := user.Current(ctx)
	if u == nil {
		fmt.Fprint(w, "false")
		return
	}
	fmt.Fprint(w, "true")
}

func getLoginAddress(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	vars, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	url, _ := user.LoginURL(ctx, fmt.Sprintf("%v", vars["url"][0]))
	fmt.Fprint(w, url)
	return
}
