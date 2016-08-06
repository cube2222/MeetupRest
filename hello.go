package hello

import (
    "fmt"
    "github.com/gorilla/mux"
    "net/http"
)

func init() {
    m := mux.NewRouter()
    m.HandleFunc("/", handler)
    http.Handle("/", m)
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Hello, world from gorilla!!!")
}
