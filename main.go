package main

import (
	"net/http"
	"os"

	"github.com/willmadison/resourceful/repository"
	resourceful "github.com/willmadison/resourceful/types"
)

func main() {
	server := resourceful.NewServer(repository.NewInMemory())
	http.Handle("/", server.Router)

	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
