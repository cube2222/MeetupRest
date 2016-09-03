package MeetupRest

import (
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"google.golang.org/appengine/aetest"
	"google.golang.org/cloud/datastore"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type speakerStoreMock struct {
	storage      map[int64]Speaker
	storageMutex sync.RWMutex
	indexCounter int64
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
	delete(s.storage, id)
	s.storageMutex.Unlock()
	return nil
}

func NewSpeakerStoreMock() *speakerStoreMock {
	return &speakerStoreMock{storage: make(map[int64]Speaker), storageMutex: sync.RWMutex{}}
}

func TestGetSpeaker(t *testing.T) {
	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Error(err)
	}
	defer inst.Close()
	router := mux.NewRouter()
	err = RegisterSpeakerRoutes(router, NewSpeakerStoreMock())
	if err != nil {
		t.Error(err)
	}
	recorder := httptest.NewRecorder()
	req, err := inst.NewRequest("GET", "http://localhost:8080/123123123/", nil)
	router.ServeHTTP(recorder, req)

	result := recorder.Result()
	data, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Error(err)
	}
	t.Log(result.StatusCode)
	t.Log(string(data))
	if result.StatusCode != http.StatusNotFound {
		t.Errorf("Nonexistent key should not be found. Wrong status. Received: %v with body: %s", result.StatusCode, data)
	}
}
