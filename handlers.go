package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/thoeni/go-tfl"
	"log"
	"net/http"
	"sort"
	"strings"
)

var tubeService = TubeService{tflClient}

func lineStatusHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	tubeLine, _ := vars["line"]

	var lines []string
	if tubeLine != "" {
		lines = []string{tubeLine}
	}

	reportsMap, err := tubeService.GetStatusFor(lines)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode("There was an error getting information from TFL"); err != nil {
			log.Panic(err)
		}
		return
	} else if len(reportsMap) == 0 {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode("Line requested not found"); err != nil {
			log.Panic(err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(reportsMap); err != nil {
		log.Panic(err)
	}
}

func slackRequestHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var slackResp slackResponse
	var slackReq = new(slackRequest)
	decoder := schema.NewDecoder()

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slackResp.Text = "There was an error parsing input data"
	} else if err := decoder.Decode(slackReq, r.PostForm); err != nil {
		println("Decoding error")
		w.WriteHeader(http.StatusBadRequest)
		slackResp.Text = "Request provided coudln't be decoded"
	} else if !isTokenValid(slackReq.Token) {
		fmt.Printf("Invalid token in request: %v from postForm: %v", slackReq, r.PostForm)
		w.WriteHeader(http.StatusUnauthorized)
		slackResp.Text = "Unauthorised"
	} else {
		tubeLine := strings.Join(slackReq.Text, " ")

		w.WriteHeader(http.StatusOK)
		slackResp.ResponseType = "ephemeral"
		slackResp.Text = fmt.Sprintf("Slack Tube Service")

		var lines []string
		if tubeLine != "" {
			lines = []string{tubeLine}
		}

		reportsMap, err := tubeService.GetStatusFor(lines)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slackResp.Text = "Error while retrieving information from TFL"
		} else if len(reportsMap) == 0 {
			w.WriteHeader(http.StatusNotFound)
			slackResp.Text = "Not a recognised line."
		}

		slackResp.Attachments = reportMapToSortedAttachmentsArray(reportsMap)
	}

	if err := json.NewEncoder(w).Encode(slackResp); err != nil {
		log.Panic(err)
	}
}

func reportMapToSortedAttachmentsArray(inputMap map[string]tfl.Report) []attachment {
	keys := make([]string, len(inputMap))
	attachments := make([]attachment, len(inputMap))
	i := 0

	for k := range inputMap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	for j, k := range keys {
		attachments[j] = mapTflLineToSlackAttachment(inputMap[k])
	}

	return attachments
}

func slackTokenRequestHandler(w http.ResponseWriter, r *http.Request) {
	token, _ := mux.Vars(r)["token"]
	switch r.Method {
	case http.MethodPut:
		err := validateToken(token)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		tokenStore.addToken(token)
	case http.MethodDelete:
		tokenStore.deleteToken(token)
	}
	w.WriteHeader(http.StatusAccepted)
}
