package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"
	"github.com/thoeni/go-tfl"
)

var tokenStore Repository

var listenPort = os.Getenv("PORT")

const defaultPort = "1123"

var (
	httpResponsesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "production",
			Subsystem: "http_server",
			Name:      "http_responses_total",
			Help:      "The count of http responses issued, classified by method and tubeLine.",
		},
		[]string{"method", "tubeLine"},
	)

	tflResponseLatencies = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "production",
			Subsystem: "tfl_client",
			Name:      "response_latencies",
			Help:      "Distribution of http response latencies (ms), classified by code and method.",
		},
	)
)

var tflClient = &InMemoryCachedClient{
	tfl.NewClient(),
	[]tfl.Report{},
	time.Now().Add(-121 * time.Second),
	float64(120),
}

var tubeService TflService = TubeService{tflClient}

func init() {

	if listenPort == "" {
		listenPort = defaultPort
	}

	err := dbInit()
	if err != nil {
		log.Fatal("Couldn't initialise DB", err)
	} else {
		fmt.Printf("BoltDB initiliased (%v), bucket created!\n", tokenStore)
	}

	prometheus.MustRegister(httpResponsesTotal)
	prometheus.MustRegister(tflResponseLatencies)
}

func main() {

	defer tokenStore.Close()

	_, authorisedTokenSet = tokenStore.RetrieveAllTokens()
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

	tokenStore = boltRepository{boltDB: db}

	return nil
}
