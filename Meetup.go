package MeetupRest

import "time"

type Meetup struct {
	Title         string
	Description   string
	Presentations []string
	Date          time.Time
	VoteTimeEnd   time.Time
}

func GetMeetupHandler() http.Handler {
	m := mux.NewRouter()
	m.HandleFunc("/meetup", getMeetup).Methods("GET")
	m.HandleFunc("/meetup", addMeetup).Methods("POST")

	return m
}
