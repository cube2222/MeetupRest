package MeetupRest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
)

type MeetupCreateData struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Time        int64   `json:"time"`
	Latitude    float64 `json:"lat"`
	Longitude   float64 `json:"lon"`
	RsvpLimit   int     `json:"self_rsvp"`
	Visibility  string  `json:"venue_visibility"`
}

type MeetupCreateResponse struct {
	ID string `json:"id"`
}

const URL = "https://api.meetup.com"

func getMeetupUpdateFunction(MetadataStorage MetadataStore, MeetupStorage MeetupStore) func(context.Context) error {
	return func(ctx context.Context) error {
		errorChan := make(chan error)
		APIKEYChan := make(chan string)
		GroupNameChan := make(chan string)
		MeetupsChan := make(chan []Meetup)
		go func() {
			APIKEY, err := MetadataStorage.GetData(ctx, "APIKEY")
			if err != nil {
				errorChan <- err
				return
			}
			APIKEYChan <- APIKEY
		}()
		go func() {
			GroupName, err := MetadataStorage.GetData(ctx, "GroupName")
			if err != nil {
				errorChan <- err
				return
			}
			GroupNameChan <- GroupName
		}()
		go func() {
			_, meetups, err := MeetupStorage.GetAllMeetups(ctx)
			if err != nil {
				errorChan <- err
				return
			}
			MeetupsChan <- meetups
		}()

		var APIKEY string
		var GroupName string
		var meetups []Meetup

		for i := 0; i < 3; i++ {
			select {
			case err := <-errorChan:
				return err
			case APIKEY = <-APIKEYChan:
			case GroupName = <-GroupNameChan:
			case meetups = <-MeetupsChan:
			}
		}

		errorChan = make(chan error)
		client := urlfetch.Client(ctx)
		for _, meetup := range meetups {
			go func(meetup Meetup) {
				Url, err := url.Parse(URL)
				if err != nil {
					errorChan <- err
					return
				}

				Url.Path += fmt.Sprintf("/%s/events/%s", GroupName, meetup.ExternalID)

				parameters := url.Values{}
				parameters = prepareMeetupDependentParams(parameters, meetup)
				parameters = prepareAuthenticationParams(parameters, APIKEY)
				Url.RawQuery = parameters.Encode()

				// TODO: Do some fun stuff. Like presentation scheduling.

				r, err := http.NewRequest("PATCH", Url.String(), nil)
				if err != nil {
					errorChan <- err
					return
				}
				_, err = client.Do(r)
				if err != nil {
					errorChan <- err
					return
				}
				errorChan <- nil
			}(meetup)
		}
		var err error
		for i := 0; i < len(meetups); i++ {
			internalErr := <-errorChan
			if internalErr != nil {
				err = internalErr
			}
		}
		if err != nil {
			return err
		}

		return nil
	}
}

func getMeetupCreateFunction(MetadataStorage MetadataStore, MeetupStorage MeetupStore) func(context.Context, int64) error {
	return func(ctx context.Context, ID int64) error {
		errorChan := make(chan error)
		APIKEYChan := make(chan string)
		GroupNameChan := make(chan string)
		MeetupChan := make(chan Meetup)
		go func() {
			APIKEY, err := MetadataStorage.GetData(ctx, "APIKEY")
			if err != nil {
				errorChan <- err
				return
			}
			APIKEYChan <- APIKEY
		}()
		go func() {
			GroupName, err := MetadataStorage.GetData(ctx, "GroupName")
			if err != nil {
				errorChan <- err
				return
			}
			GroupNameChan <- GroupName
		}()
		go func() {
			meetup, err := MeetupStorage.GetMeetup(ctx, ID)
			if err != nil {
				errorChan <- err
				return
			}
			MeetupChan <- meetup
		}()

		var APIKEY string
		var GroupName string
		var meetup Meetup

		for i := 0; i < 3; i++ {
			select {
			case err := <-errorChan:
				return err
			case APIKEY = <-APIKEYChan:
			case GroupName = <-GroupNameChan:
			case meetup = <-MeetupChan:
			}
		}

		Url, err := url.Parse(URL)
		if err != nil {
			return err
		}

		Url.Path += fmt.Sprintf("/%s/events", GroupName)

		parameters := url.Values{}
		parameters = prepareMeetupDependentParams(parameters, meetup)
		parameters = prepareAuthenticationParams(parameters, APIKEY)
		Url.RawQuery = parameters.Encode()

		client := urlfetch.Client(ctx)
		res, err := client.Post(Url.String(), "", nil) // this header is necessary?
		if err != nil {
			return err
		}
		meetupResponse := MeetupCreateResponse{}
		err = json.NewDecoder(res.Body).Decode(&meetupResponse)
		if err != nil {
			return err
		}

		meetup.ExternalID = meetupResponse.ID
		err = MeetupStorage.PutMeetup(ctx, ID, &meetup)
		return err
	}
}

func prepareMeetupDependentParams(parameters url.Values, meetup Meetup) url.Values {
	parameters.Add("name", meetup.Title)
	parameters.Add("description", meetup.Description)
	parameters.Add("time", fmt.Sprintf("%v", meetup.Date.UnixNano()/int64(time.Millisecond)))
	parameters.Add("lat", fmt.Sprintf("%v", meetup.Latitude))
	parameters.Add("lon", fmt.Sprintf("%v", meetup.Longitude))
	parameters.Add("venue_visibility", "members")

	return parameters
}

func prepareAuthenticationParams(parameters url.Values, APIKey string) url.Values {
	parameters.Add("sign", "true")
	parameters.Add("key", APIKey)

	return parameters
}
