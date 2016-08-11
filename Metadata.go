package MeetupRest

import (
	"context"
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
	keyId := datastore.NewKey(ctx, datastoreMetadataKind, key, 0, nil)
	if _, err := datastore.Put(ctx, keyId, myMetadata); err != nil {
		log.Errorf(ctx, "Can't create datastore object: %v", err)
	}
}

func getData(key string) string {
	ctx := context.Background()
	var data Metadata
	keyId := datastore.NewKey(ctx, datastoreMetadataKind, key, 0, nil)
	if err := datastore.Get(ctx, keyId, &data); err != nil {
		log.Errorf(ctx, "Can't retrive datastore object", err)
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
