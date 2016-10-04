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
	"github.com/thoeni/go-tfl"
)

const minStatusPollPeriod = 2

var statuses []tfl.Report

func lineStatusHandler(w http.ResponseWriter, r *http.Request) {

	var response []tfl.Report

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
			response = append(response, line)
		}
	} else {
		for _, line := range statuses {
			if strings.ToLower(line.Name) == strings.ToLower(tubeLine) {
				response = append(response, line)
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

	var slackResp slackResponse
	var attachments []attachment
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
		if isUpdateNeeded() {
			if err := updateStatusInformation(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				slackResp.Text = "There was an error getting information from TFL"
			}
		}

		tubeLine := strings.Join(slackReq.Text, " ")

		w.WriteHeader(http.StatusOK)
		slackResp.ResponseType = "ephemeral"
		slackResp.Text = fmt.Sprintf("Slack Tube Service - last updated at %s", lastStatusCheck.Format("15:04:05"))

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
				slackResp.Text = "Not a recognised line."
			}
		}

		slackResp.Attachments = attachments
	}

	if err := json.NewEncoder(w).Encode(slackResp); err != nil {
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
		tokenStore.addSlackToken(token)
	case http.MethodDelete:
		tokenStore.deleteSlackToken(token)
	}
	w.WriteHeader(http.StatusAccepted)
}

func isUpdateNeeded() bool {
	return time.Since(lastStatusCheck).Minutes() > minStatusPollPeriod
}

func updateStatusInformation() error {
	client := tfl.NewClient()
	reports, err := client.GetTubeStatus()
	if err != nil {
		log.Print("Error while retrieving Tube statuses")
		return err
	}
	statuses = reports
	lastStatusCheck = time.Now()
	return nil
}
