package MeetupRest

import (
	"net/http"

	"fmt"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/user"
)

func init() {

	m := mux.NewRouter()

	s := m.PathPrefix("/speaker").Subrouter()
	err := RegisterSpeakerRoutes(s)

	s = m.PathPrefix("/presentation").Subrouter()
	err = RegisterPresentationRoutes(s)

	s = m.PathPrefix("/meetup").Subrouter()
	err = RegisterMeetupRoutes(s)

	m.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("public/"))))

	m.HandleFunc("/isLoggedIn", isLoggedIn)

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
