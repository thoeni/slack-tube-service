package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/thoeni/go-tfl"
	"net/http"
	"sort"
	"strings"
)

func slackRequestHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	var slackResp slackResponse
	var slackReq = new(slackRequest)
	decoder := schema.NewDecoder()

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slackResp.Text = "There was an error parsing input data"
		encoder.Encode(slackResp)
		return
	} else if err := decoder.Decode(slackReq, r.PostForm); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		slackResp.Text = "Request provided coudln't be decoded"
		encoder.Encode(slackResp)
		return
	} else if !isTokenValid(slackReq.Token) {
		fmt.Printf("Invalid token in request: %v from postForm: %v", slackReq, r.PostForm)
		w.WriteHeader(http.StatusUnauthorized)
		slackResp.Text = "Unauthorised"
		encoder.Encode(slackResp)
		return
	}

	tubeLine := strings.Join(slackReq.Text, " ")

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
		encoder.Encode(slackResp)
		return
	} else if len(reportsMap) == 0 {
		w.WriteHeader(http.StatusOK)
		slackResp.Text = "Not a recognised line."
		encoder.Encode(slackResp)
		return
	}

	w.WriteHeader(http.StatusOK)
	slackResp.Attachments = reportMapToSortedAttachmentsArray(reportsMap)
	encoder.Encode(slackResp)
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
		tokenStore.AddToken(token)
	case http.MethodDelete:
		tokenStore.DeleteToken(token)
	}
	w.WriteHeader(http.StatusAccepted)
}
