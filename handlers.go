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
	"github.com/thoeni/slack-tube-service/service"
)

const minStatusPollPeriod = 2

var statuses []tfl.Report
var reportMap map[string]tfl.Report
var client tfl.Client = service.InMemoryCachedClient{
	tfl.NewClient(),
	[]tfl.Report{},
	time.Now().Add(-121 * time.Second),
	float64(120),
}
var tubeService = service.TubeService{client}

func lineStatusHandler(w http.ResponseWriter, r *http.Request) {

	//var response []tfl.Report

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	tubeLine, _ := vars["line"]

	log.Printf("Line requested: %s", tubeLine)

	status, err := tubeService.GetStatusFor([]string{tubeLine})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode("There was an error getting information from TFL"); err != nil {
			log.Panic(err)
		}
	}

	log.Printf("Status size is: %d", len(status))

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Panic(err)
	}
}

func lineStatusHandler2(w http.ResponseWriter, r *http.Request) {

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
		response = statuses
	} else {
		report, found := reportMap[strings.ToLower(tubeLine)]
		if !found {
			w.WriteHeader(http.StatusNotFound)
			if err := json.NewEncoder(w).Encode("Not a recognised line."); err != nil {
				log.Panic(err)
			}
			return
		}
		response = append(response, report)
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

func updateStatusInformation() error {
	var client tfl.Client = tfl.NewClient()
	reports, err := client.GetTubeStatus()
	if err != nil {
		log.Print("Error while retrieving Tube statuses")
		return err
	}
	statuses = reports
	reportMap = tfl.ReportArrayToMap(reports)
	lastStatusCheck = time.Now()
	return nil
}

func isUpdateNeeded() bool {
	return time.Since(lastStatusCheck).Minutes() > minStatusPollPeriod
}

func getReports() ([]tfl.Report, error) {
	if isUpdateNeeded() {
		if err := updateStatusInformation(); err != nil {
			return nil, err
		}
	}
	return statuses, nil
}
