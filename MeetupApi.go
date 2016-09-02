package MeetupRest

type MeetupCreateData struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Time        int64   `json:"time"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	RsvpLimit   int     `json:"rsvp_limit"`
	Visibility  string  `json:"visibility"`
}

func getMeetupUpdateFunction(Storage *MetadataStore) func() error {
	return func() error {
		// Here we can use the Storage
		return nil
	}
}
