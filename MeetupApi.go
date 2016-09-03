package MeetupRest

import (
	"bytes"
	"encoding/json"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"io/ioutil"
	"net/url"
)

type MeetupCreateData struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Time        int64   `json:"time"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	RsvpLimit   int     `json:"self_rsvp"`
	Visibility  string  `json:"venue_visibility"`
}

func getMeetupUpdateFunction(MetadataStorage MetadataStore, MeetupStorage MeetupStore) func(context.Context) error {
	return func(ctx context.Context) error {
		return nil
	}
}

func getMeetupCreateFunction(MetadataStorage MetadataStore, MeetupStorage MeetupStore) func(context.Context, int64) error {
	return func(ctx context.Context, ID int64) error {
		APIKEY, err := MetadataStorage.GetData(ctx, "APIKEY")
		if err != nil {
			return err
		}
		meetup, err := MeetupStorage.GetMeetup(ctx, ID)
		if err != nil {
			return err
		}

		mcd := MeetupCreateData{
			Name:        meetup.Title,
			Description: meetup.Description,
			Time:        meetup.Date.Unix(),
			Lat:         0.0,
			Lon:         0.0,
			RsvpLimit:   25,
			Visibility:  "members",
		}
		data, err := json.Marshal(mcd)
		if err != nil {
			return nil
		}
		log.Infof(ctx, "Data to send to Meetup.com: %s", data)
		client := urlfetch.Client(ctx)
		Url, err := url.Parse("https://api.meetup.com")
		if err != nil {
			return err
		}
		// TODO: Golang-Warsaw in metadata
		Url.Path += "/Golang-Warsaw/events"
		parameters := url.Values{}
		parameters.Add("name", mcd.Name)
		parameters.Add("sign", "true")
		parameters.Add("key", APIKEY)
		Url.RawQuery = parameters.Encode()
		res, err := client.Post(Url.String(), "application/json", bytes.NewBuffer(data))
		if err != nil {
			return err
		}
		log.Infof(ctx, "%v", res.StatusCode)
		data, _ = ioutil.ReadAll(res.Body)
		log.Infof(ctx, "%s", data)
		return nil
	}
}
