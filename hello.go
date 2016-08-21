package MeetupRest

import (
	"net/http"

	"github.com/gorilla/mux"
)

func init() {

	m := mux.NewRouter()

	s := m.PathPrefix("/speaker").Subrouter()
	err := RegisterSpeakerRoutes(s)

	s = m.PathPrefix("/presentation").Subrouter()
	err = RegisterPresentationRoutes(s)

	s = m.PathPrefix("/meetup").Subrouter()
	err = RegisterMeetupRoutes(s)

	if err != nil {
		panic(err)
	}

	http.Handle("/", m)
}
