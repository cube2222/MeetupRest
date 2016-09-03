package MeetupRest

import "golang.org/x/net/context"

type MeetupCreateData struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Time        int64   `json:"time"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	RsvpLimit   int     `json:"rsvp_limit"`
	Visibility  string  `json:"visibility"`
}

func getMeetupUpdateFunction(MetadataStorage MetadataStore, MeetupStorage MeetupStore) func(context.Context) error {
	return func(ctx context.Context) error {
		// Here we can use the Storage
		return nil
	}
}

func getMeetupCreateFunction(MetadataStorage MetadataStore, MeetupStorage MeetupStore) func(context.Context, string) error {
	return func(ctx context.Context, Name string) error {
		// Here we can use the Storage
		return nil
	}
}
