package MeetupRest

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func GetFormsHandler() http.Handler {
	m := mux.NewRouter()
	m.HandleFunc("/addSpeaker", addSpeakerForm).Methods("GET")

	return m
}

func addSpeakerForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Adding Speaker Form</h1>"+
		"<form action=\"/speaker\" method=\"POST\">"+
		"First name: <input type=\"text\" name=\"Name\"><br>"+
		"Last name: <input type=\"text\" name=\"Surname\"><br>"+
		"Company: <input type=\"text\" name=\"Company\"><br>"+
		"email: <input type=\"email\" name=\"Email\"><br>"+
		"About: <textarea name=\"About\"></textarea><br>"+
		"<input type=\"submit\" value=\"Save\">"+
		"</form>")
}
