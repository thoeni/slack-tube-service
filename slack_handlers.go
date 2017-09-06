package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"net/http"
	"strings"
)

func slackRequestHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	var slackResp slackResponse = NewEphemeral()
	var slackReq = new(slackRequest)
	decoder := schema.NewDecoder()

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slackResp.Text = "There was an error parsing input data"
		encoder.Encode(slackResp)
		return
	}

	if err := decoder.Decode(slackReq, r.PostForm); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		slackResp.Text = "Request provided coudln't be decoded"
		encoder.Encode(slackResp)
		return
	}

	if !isTokenValid(slackReq.Token) {
		fmt.Printf("Invalid token in request: %v from postForm: %v", slackReq, r.PostForm)
		w.WriteHeader(http.StatusUnauthorized)
		slackResp.Text = "Unauthorised"
		encoder.Encode(slackResp)
		return
	}

	if len(slackReq.Text) == 0 {
		sr := NewEphemeral()
		sr.Text = fmt.Sprint("This slack integration provides four options:\n\n-`/tube status` or `/tube status <lineName>` for example `/tube status bakerloo`\n-`/tube subscribe <lineName>`, for example `/tube subscribe bakerloo`\n-`/tube for <username>`, for example `/tube for @jlennon`\n-`/tube version`\n\nFor more details please visit <https://thoeni.io/project/slack-tube-service/|slack-tube-service project page>")
		w.WriteHeader(http.StatusOK)
		encoder.Encode(sr)
		return
	}
	slackInput := strings.Split(slackReq.Text[0], " ")

	slackCommand := slackInput[0]
	slackCommandArgs := slackInput[1:]

	switch slackCommand {
	case "status":
		sr, _ := statusCommand(slackCommandArgs, *slackReq)
		w.WriteHeader(http.StatusOK)
		encoder.Encode(sr)
	case "for":
		sr, _ := forCommand(slackCommandArgs, *slackReq)
		w.WriteHeader(http.StatusOK)
		encoder.Encode(sr)
	case "subscribe":
		sr, _ := subscribeCommand(slackCommandArgs, *slackReq)
		w.WriteHeader(http.StatusOK)
		encoder.Encode(sr)
	case "version":
		sr, _ := versionCommand(slackCommandArgs, *slackReq)
		w.WriteHeader(http.StatusOK)
		encoder.Encode(sr)
	default:
		sr := NewEphemeral()
		sr.Text = fmt.Sprintf("Unrecognised command: %s", slackCommand)
		w.WriteHeader(http.StatusOK)
		encoder.Encode(sr)
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
		tokenStore.AddToken(token)
	case http.MethodDelete:
		tokenStore.DeleteToken(token)
	}
	w.WriteHeader(http.StatusAccepted)
}
