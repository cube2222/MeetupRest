package MeetupRest

import "time"

type Meetup struct {
	Title         string
	Description   string
	Presentations []string
	Date          time.Time
	VoteTimeEnd   time.Time
}

type Presentation struct {
	Title       string
	Description string
	Author      string
	VoteCount   int
}

type Author struct {
	Name          string
	Surname       string
	About         string
	Email         string
	Company       string
	Presentations []string
}
