package MeetupRest

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type GoogleDatastoreStore struct {
}

type data struct {
	Content string
}

func (ds *GoogleDatastoreStore) GetSpeaker(ctx context.Context, ID int64) (Speaker, error) {
	speaker := Speaker{}
	key := datastore.NewKey(ctx, datastoreSpeakersKind, "", ID, nil)
	err := datastore.Get(ctx, key, &speaker)
	return speaker, err
}

func (ds *GoogleDatastoreStore) GetAllSpeakers(ctx context.Context) ([]int64, []Speaker, error) {
	speakers := make([]Speaker, 0, 10)
	keys, err := datastore.NewQuery(datastoreSpeakersKind).GetAll(ctx, &speakers)

	IDs := make([]int64, 0, len(speakers))
	for _, key := range keys {
		IDs = append(IDs, key.IntID())
	}

	return IDs, speakers, err
}

func (ds *GoogleDatastoreStore) PutSpeaker(ctx context.Context, ID int64, speaker *Speaker) error {
	key := datastore.NewKey(ctx, datastoreSpeakersKind, "", ID, nil)
	_, err := datastore.Put(ctx, key, speaker)
	return err
}

func (ds *GoogleDatastoreStore) AddSpeaker(ctx context.Context, speaker *Speaker) (int64, error) {
	key := datastore.NewKey(ctx, datastoreSpeakersKind, "", 0, nil)
	ID, err := datastore.Put(ctx, key, speaker)
	return ID.IntID(), err
}

func (ds *GoogleDatastoreStore) DeleteSpeaker(ctx context.Context, ID int64) error {
	key := datastore.NewKey(ctx, datastoreSpeakersKind, "", ID, nil)
	return datastore.Delete(ctx, key)
}

func (ds *GoogleDatastoreStore) GetPresentation(ctx context.Context, ID int64) (Presentation, error) {
	presentation := Presentation{}
	key := datastore.NewKey(ctx, datastorePresentationsKind, "", ID, nil)
	err := datastore.Get(ctx, key, &presentation)
	return presentation, err
}

func (ds *GoogleDatastoreStore) GetAllPresentations(ctx context.Context) ([]int64, []Presentation, error) {
	presentations := make([]Presentation, 0, 10)
	keys, err := datastore.NewQuery(datastorePresentationsKind).GetAll(ctx, &presentations)

	IDs := make([]int64, 0, len(presentations))
	for _, key := range keys {
		IDs = append(IDs, key.IntID())
	}

	return IDs, presentations, err
}

func (ds *GoogleDatastoreStore) PutPresentation(ctx context.Context, ID int64, presentation *Presentation) error {
	key := datastore.NewKey(ctx, datastorePresentationsKind, "", ID, nil)
	_, err := datastore.Put(ctx, key, presentation)
	return err
}

func (ds *GoogleDatastoreStore) AddPresentation(ctx context.Context, presentation *Presentation) (int64, error) {
	key := datastore.NewKey(ctx, datastorePresentationsKind, "", 0, nil)
	ID, err := datastore.Put(ctx, key, presentation)
	return ID.IntID(), err
}

func (ds *GoogleDatastoreStore) DeletePresentation(ctx context.Context, ID int64) error {
	key := datastore.NewKey(ctx, datastorePresentationsKind, "", ID, nil)
	return datastore.Delete(ctx, key)
}

func (ds *GoogleDatastoreStore) GetMeetup(ctx context.Context, ID int64) (Meetup, error) {
	meetup := Meetup{}
	key := datastore.NewKey(ctx, datastoreMeetupsKind, "", ID, nil)
	err := datastore.Get(ctx, key, &meetup)
	return meetup, err
}

func (ds *GoogleDatastoreStore) GetAllMeetups(ctx context.Context) ([]int64, []Meetup, error) {
	meetups := make([]Meetup, 0, 10)
	keys, err := datastore.NewQuery(datastoreMeetupsKind).GetAll(ctx, &meetups)

	IDs := make([]int64, 0, len(meetups))
	for _, key := range keys {
		IDs = append(IDs, key.IntID())
	}

	return IDs, meetups, err
}

func (ds *GoogleDatastoreStore) PutMeetup(ctx context.Context, ID int64, meetup *Meetup) error {
	key := datastore.NewKey(ctx, datastoreMeetupsKind, "", ID, nil)
	_, err := datastore.Put(ctx, key, meetup)
	return err
}

func (ds *GoogleDatastoreStore) AddMeetup(ctx context.Context, meetup *Meetup) (int64, error) {
	key := datastore.NewKey(ctx, datastoreMeetupsKind, "", 0, nil)
	ID, err := datastore.Put(ctx, key, meetup)
	return ID.IntID(), err
}

func (ds *GoogleDatastoreStore) DeleteMeetup(ctx context.Context, ID int64) error {
	key := datastore.NewKey(ctx, datastoreMeetupsKind, "", ID, nil)
	return datastore.Delete(ctx, key)
}

func (ds *GoogleDatastoreStore) GetData(ctx context.Context, key string) (string, error) {
	data := data{}
	keyInternal := datastore.NewKey(ctx, datastoreMetadataKind, key, 0, nil)
	err := datastore.Get(ctx, keyInternal, &data)
	return data.Content, err
}

func (ds *GoogleDatastoreStore) PutData(ctx context.Context, key string, value string) error {
	dataInternal := data{Content: value}
	keyInternal := datastore.NewKey(ctx, datastoreMetadataKind, key, 0, nil)
	_, err := datastore.Put(ctx, keyInternal, &dataInternal)
	return err
}

func (ds *GoogleDatastoreStore) DeleteData(ctx context.Context, key string) error {
	keyInternal := datastore.NewKey(ctx, datastoreMetadataKind, key, 0, nil)
	err := datastore.Delete(ctx, keyInternal)
	return err
}
