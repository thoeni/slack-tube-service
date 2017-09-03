package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/boltdb/bolt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"
	"github.com/thoeni/go-tfl"
)

var tokenStore Repository
var svc *dynamodb.DynamoDB

var listenPort = os.Getenv("PORT")
var AppVersion string
var Sha string

const defaultPort = "1123"

var (
	httpResponsesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "production",
			Subsystem: "http_server",
			Name:      "http_requests_total",
			Help:      "The count of http responses issued, classified by method and requestURI.",
		},
		[]string{"method", "requestURI"},
	)

	slackRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "production",
			Subsystem: "http_server",
			Name:      "slack_requests_total",
			Help:      "The count of http requests received for tube status, classified by slackDomain and tubeLine.",
		},
		[]string{"domain", "tubeLine"},
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

var tubeService TflService = TubeService{tfl.NewCachedClient(120)}

func initialise() {

	if listenPort == "" {
		listenPort = defaultPort
	}

	err := dbInit()
	if err != nil {
		log.Fatal("Couldn't initialise DB", err)
		return
	}
	fmt.Printf("BoltDB initiliased (%v), bucket created!\n", tokenStore)

	// DynamoDB
	sess := session.Must(session.NewSession())
	svc = dynamodb.New(sess, aws.NewConfig().WithRegion("eu-west-1"))

	// Prometheus
	prometheus.MustRegister(httpResponsesTotal)
	prometheus.MustRegister(slackRequestsTotal)
	prometheus.MustRegister(tflResponseLatencies)
}

func main() {

	printVersion := flag.Bool("version", false, "Prints the version of this application")
	flag.Parse()
	if *printVersion {
		fmt.Printf("Current version is: %s\nGit commit: %s", AppVersion, Sha)
		return
	}

	initialise()
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
