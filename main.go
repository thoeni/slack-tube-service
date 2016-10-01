package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rs/cors"
	"os"
)

var lastStatusCheck time.Time

var listenPort = os.Getenv("PORT")

const defaultPort = "1123"

func init() {

	if listenPort == "" {
		listenPort = defaultPort
	}

	var err error
	lastStatusCheck, err = time.Parse(time.RFC3339, "1970-01-01T00:00:00+00:00")
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	loadAuthorisedTokensFromFile(authorisedTokenFileLocation)
	router := newRouter()
	fmt.Println("Ready, listening on port", listenPort)
	log.Fatal(http.ListenAndServe(":"+listenPort, cors.Default().Handler(router)))
}
