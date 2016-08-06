package MeetupRest

import "time"

type Meetup struct {
	Title         string
	Description   string
	Presentations []string
	Date          time.Time
	VoteTimeEnd   time.Time
}
