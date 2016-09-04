package MeetupRest

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
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

const URL = "https://api.meetup.com"

func getMeetupUpdateFunction(MetadataStorage MetadataStore, MeetupStorage MeetupStore) func(context.Context) error {
	return func(ctx context.Context) error {
		APIKEY, err := MetadataStorage.GetData(ctx, "APIKEY")
		if err != nil {
			return err
		}

		_, meetups, err := MeetupStorage.GetAllMeetups(ctx)
		if err != nil {
			return err
		}

		for _, meetup := range meetups {
			client := urlfetch.Client(ctx)
			Url, err := url.Parse(URL)
			if err != nil {
				return err
			}
			Url.Path += "/Golang-Warsaw/events/" + meetup.EventId
			parameters := prepareParamsUrl(meetup, APIKEY)
			Url.RawQuery = parameters.Encode()
			log.Infof(ctx, Url.String())

			// this header is necessary?, How can I do PATH method?
			res, err := client.Post(Url.String(), "application/json", nil)
			if err != nil {
				return err
			}
			log.Infof(ctx, "%v", res.StatusCode)
		}

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

		client := urlfetch.Client(ctx)
		Url, err := url.Parse(URL)
		if err != nil {
			return err
		}
		// TODO: Golang-Warsaw in metadata
		Url.Path += "/Golang-Warsaw/events"
		parameters := prepareParamsUrl(meetup, APIKEY)
		Url.RawQuery = parameters.Encode()
		log.Infof(ctx, Url.String())
		res, err := client.Post(Url.String(), "application/json", nil) // this header is necessary?
		if err != nil {
			return err
		}
		log.Infof(ctx, "%v", res.StatusCode)
		data, _ := ioutil.ReadAll(res.Body)
		log.Infof(ctx, "%s", data)
		// TODO: Extract from response 'eventId' and update 'meetup' in datastore!

		return nil
	}
}

func prepareParamsUrl(meetup Meetup, apiKey string) url.Values {
	parameters := url.Values{}
	parameters.Add("name", meetup.Title)
	parameters.Add("description", meetup.Description)
	parameters.Add("time", fmt.Sprintf("%v", meetup.Date.UnixNano()/int64(time.Millisecond)))
	parameters.Add("lat", fmt.Sprintf("%v", meetup.Lat))
	parameters.Add("lon", fmt.Sprintf("%v", meetup.Lon))
	parameters.Add("venue_visibility", "members")
	parameters.Add("sign", "true")
	parameters.Add("key", apiKey)

	return parameters
}
