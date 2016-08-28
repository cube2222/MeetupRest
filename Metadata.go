package MeetupRest

import (
	"golang.org/x/net/context"
	"time"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const datastoreMetadataKind = "Metadata"
const meetupAPI = "MEETUP_API"

type Metadata struct {
	Key     string
	Content string
}

func setData(key, content string) {
	ctx := context.Background()
	myMetadata := &Metadata{
		Key:     key,
		Content: content,
	}
	keyID := datastore.NewKey(ctx, datastoreMetadataKind, key, 0, nil)
	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	if _, err := datastore.Put(newCtx, keyID, myMetadata); err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
	}
}

func getData(key string) string {
	ctx := context.Background()
	var data Metadata
	keyID := datastore.NewKey(ctx, datastoreMetadataKind, key, 0, nil)

	newCtx, _ := context.WithTimeout(ctx, time.Second*2)
	if err := datastore.Get(newCtx, keyID, &data); err != nil {
		log.Errorf(ctx, "Can't retrive datastore object: %v", err)
		return ""
	}
	return data.Content
}

func (m Metadata) GetMeetupAPI() string {
	return getData(meetupAPI)
}

func (m *Metadata) SetMeetupAPI(content string) {
	setData(meetupAPI, content)
}
