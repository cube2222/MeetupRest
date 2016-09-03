package MeetupRest

import (
	"golang.org/x/net/context"
	"google.golang.org/cloud/datastore"
	"sync"
	"testing"
)

type speakerStoreMock struct {
	storage      map[int64]Speaker
	storageMutex sync.RWMutex
	indexCounter int
}

func (s *speakerStoreMock) GetSpeaker(ctx context.Context, id int64) (Speaker, error) {
	s.storageMutex.RLock()
	speaker, ok := s.storage[id]
	s.storageMutex.RUnlock()
	if !ok {
		return speaker, datastore.ErrNoSuchEntity
	} else {
		return speaker, nil
	}
}

func (s *speakerStoreMock) GetAllSpeakers(ctx context.Context) ([]int64, []Speaker, error) {
	s.storageMutex.RLock()
	speakers := make([]Speaker, 0, len(s.storage))
	IDs := make([]int64, 0, len(s.storage))
	for key, value := range s.storage {
		IDs = append(IDs, key)
		speakers = append(speakers, value)
	}
	s.storageMutex.RUnlock()
	return IDs, speakers, nil
}

func (s *speakerStoreMock) PutSpeaker(ctx context.Context, id int64, speaker *Speaker) error {
	s.storageMutex.Lock()
	s.storage[id] = *speaker
	s.storageMutex.Unlock()
	return nil
}

func (s *speakerStoreMock) AddSpeaker(ctx context.Context, speaker *Speaker) (int64, error) {
	s.storageMutex.Lock()
	for _, ok := s.storage[s.indexCounter]; ok; s.indexCounter++ {
	}
	s.storage[s.indexCounter] = *speaker
	s.indexCounter++
	s.storageMutex.Unlock()
	return s.indexCounter - 1, nil
}

func (s *speakerStoreMock) DeleteSpeaker(ctx context.Context, id int64) error {
	s.storageMutex.Lock()
	delete(s.storage, s.storage[id])
	s.storageMutex.Unlock()
	return nil
}

func TestgetSpeaker(t *testing.T) {
}
