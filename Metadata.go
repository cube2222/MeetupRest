package MeetupRest

import (
	"golang.org/x/net/context"

	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
	"net/http"
	"net/url"
)

const datastoreMetadataKind = "Metadata"
const meetupAPI = "MEETUP_API"

type MetadataStore interface {
	GetData(ctx context.Context, key string) (string, error)
	PutData(ctx context.Context, key string, data string) error
	DeleteData(ctx context.Context, key string) error
}

// Register meetup routes to the router
func RegisterMetadataRoutes(m *mux.Router, Storage MetadataStore) error {
	if m == nil {
		return errors.New("m may not be nil when regitering meetup routes")
	}
	h := metadataHandler{Storage: Storage}
	m.HandleFunc("/{key}/", h.getData).Methods("GET")
	m.HandleFunc("/{key}/", h.setData).Methods("POST")

	return nil
}

type metadataHandler struct {
	Storage MetadataStore
}

func (h *metadataHandler) getData(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, fmt.Sprint("/"))
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
	}
	if !u.Admin {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "You have to be admin.")
		return
	}

	vars := mux.Vars(r)

	value, err := h.Storage.GetData(ctx, vars["key"])
	if err == datastore.ErrNoSuchEntity {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Couldn't find data with key: %v", vars["key"])
		return
	}
	if err != nil {
		log.Errorf(ctx, "Couldn't get data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, value)
}

func (h *metadataHandler) setData(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx, done := context.WithTimeout(ctx, defaultRequestTimeout)
	defer done()

	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, fmt.Sprint("/"))
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
	}
	if !u.Admin {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "You have to be admin.")
		return
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Errorf(ctx, "Error when parsing query:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, ok := params["data"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Data is mandatory.")
		return
	}

	vars := mux.Vars(r)

	err = h.Storage.PutData(ctx, vars["key"], data[0])
	if err == datastore.ErrNoSuchEntity {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Couldn't find data with key: %v", vars["key"])
		return
	}
	if err != nil {
		log.Errorf(ctx, "Couldn't get data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Successful.")
}
