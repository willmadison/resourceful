package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	router := registerEndpoints()
	http.Handle("/", router)

	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}

func registerEndpoints() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/inbox", func(response http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(response, "Hello from resourceful!")
	}).Methods("POST")

	return r
}
