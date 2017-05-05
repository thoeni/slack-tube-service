package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

func lineStatusHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	vars := mux.Vars(r)
	tubeLine, _ := vars["line"]

	var lines []string
	if tubeLine != "" {
		lines = []string{tubeLine}
	}

	reportsMap, err := tubeService.GetStatusFor(lines)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode("There was an error getting information from TFL")
		return
	} else if len(reportsMap) == 0 {
		w.WriteHeader(http.StatusNotFound)
		encoder.Encode("Line requested not found")
		return
	}

	w.WriteHeader(http.StatusOK)
	encoder.Encode(reportsMap)
}
