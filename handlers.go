package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

const minStatusPollPeriod = 2

var statuses []Report

func lineStatusHandler(w http.ResponseWriter, r *http.Request) {

	var response []Report

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if isUpdateNeeded() {
		if err := updateStatusInformation(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode("There was an error getting information from TFL"); err != nil {
				log.Panic(err)
			}
		}
	}

	vars := mux.Vars(r)
	tubeLine, lineIsPresentInPath := vars["line"]

	if !lineIsPresentInPath {
		for _, line := range statuses {
			response = append(response, mapTflLineToResponse(line))
		}
	} else {
		for _, line := range statuses {
			if strings.ToLower(line.Name) == strings.ToLower(tubeLine) {
				response = append(response, mapTflLineToResponse(line))
			}

		}
		if len(response) == 0 {
			w.WriteHeader(http.StatusNotFound)
			if err := json.NewEncoder(w).Encode("Not a recognised line."); err != nil {
				log.Panic(err)
			}
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Panic(err)
	}
}

func slackRequestHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var slackResponse SlackResponse
	var attachments []Attachment
	var slackRequest = new(SlackRequest)
	decoder := schema.NewDecoder()

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slackResponse.Text = "There was an error parsing input data"
	} else if err := decoder.Decode(slackRequest, r.PostForm); err != nil {
		println("Decoding error")
		w.WriteHeader(http.StatusBadRequest)
		slackResponse.Text = "Request provided coudln't be decoded"
	} else if !isTokenValid(slackRequest.Token) {
		println("Invalid token")
		w.WriteHeader(http.StatusUnauthorized)
		slackResponse.Text = "Unauthorised"
	} else {
		if isUpdateNeeded() {
			if err := updateStatusInformation(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				slackResponse.Text = "There was an error getting information from TFL"
			}
		}

		tubeLine := strings.Join(slackRequest.Text, " ")

		w.WriteHeader(http.StatusOK)
		slackResponse.Response_type = "ephemeral"
		slackResponse.Text = fmt.Sprintf("Slack Tube Service - last updated at %s", lastStatusCheck.Format("15:04:05"))

		if tubeLine == "" {
			for _, line := range statuses {
				attachments = append(attachments, mapTflLineToSlackAttachment(line))
			}
		} else {
			for _, line := range statuses {
				if strings.ToLower(line.Name) == strings.ToLower(tubeLine) {
					attachments = append(attachments, mapTflLineToSlackAttachment(line))
				}
			}
			if len(attachments) == 0 {
				w.WriteHeader(http.StatusNotFound)
				slackResponse.Text = "Not a recognised line."
			}
		}

		slackResponse.Attachments = attachments
	}

	if err := json.NewEncoder(w).Encode(slackResponse); err != nil {
		log.Panic(err)
	}
}

func slackTokenRequestHandler(w http.ResponseWriter, r *http.Request) {
	token, _ := mux.Vars(r)["token"]
	err := validateToken(token)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPut:
		addSlackToken(token)
	case http.MethodDelete:
		deleteSlackToken(token)
	}
	w.WriteHeader(http.StatusAccepted)
}

func isUpdateNeeded() bool {
	return time.Since(lastStatusCheck).Minutes() > minStatusPollPeriod
}

func updateStatusInformation() error {
	url := "https://api.tfl.gov.uk/line/mode/tube/status"

	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)

	var data []Report
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	statuses = data
	lastStatusCheck = time.Now()
	return nil
}
