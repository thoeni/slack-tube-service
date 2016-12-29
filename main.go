package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/rs/cors"
)

var tokenStore tokenStorer
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

	err = dbInit()
	if err != nil {
		log.Fatal("Couldn't initialise DB", err)
	} else {
		fmt.Printf("BoltDB initiliased (%v), bucket created!\n", tokenStore)
	}
}

func main() {

	defer tokenStore.close()

	_, authorisedTokenSet = tokenStore.retrieveAllTokens()
	router := newRouter()
	fmt.Println("Ready, listening on port", listenPort)
	log.Fatal(http.ListenAndServe(":"+listenPort, cors.Default().Handler(router)))
}

func dbInit() error {

	db, err := bolt.Open("slack-tube-service.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("token"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	}); err != nil {
		return err
	}

	tokenStore = boltTokenStore{boltDB: db}

	return nil
}
