package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type route struct {
	Name        string
	Methods     []string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type routes []route

func withHTTPMetricsFor(f http.HandlerFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		go func() {
			httpResponsesTotal.WithLabelValues(r.Method, r.RequestURI).Inc()
		}()
		f(w, r)
	}
}

func newRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range r {
		router.
			Methods(route.Methods...).
			Path(route.Pattern).
			Name(route.Name).
			HandlerFunc(route.HandlerFunc)
	}

	// Additional route for Prometheus instrumentation
	router.Methods(http.MethodGet).Path("/metrics").Handler(promhttp.Handler())

	return router
}

var r = routes{
	route{
		"get-all-lines-status",
		[]string{http.MethodGet},
		"/api/tubestatus/",
		withHTTPMetricsFor(lineStatusHandler),
	},
	route{
		"get-line-status",
		[]string{http.MethodGet},
		"/api/tubestatus/{line}",
		withHTTPMetricsFor(lineStatusHandler),
	},
	route{
		"slack-get-all-lines-status",
		[]string{http.MethodPost},
		"/api/slack/tubestatus/",
		withHTTPMetricsFor(slackRequestHandler),
	},
	route{
		"slack-add-auth-token",
		[]string{http.MethodPut, http.MethodDelete},
		"/api/slack/token/{token}",
		withHTTPMetricsFor(slackTokenRequestHandler),
	},
	route{
		"slack-subscribe-to-line",
		[]string{http.MethodPost},
		"/api/slack/subscribe/",
		subscribeToLine,
	},
}
