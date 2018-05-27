package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"github.com/aws/aws-lambda-go/events"
	"github.com/pkg/errors"
	"net/url"
)

func slackRequestHandler(l *tubeServuceLambda, verb string, v url.Values) (events.APIGatewayProxyResponse, error) {

	resp := events.APIGatewayProxyResponse{
		Headers: map[string]string{"Content-Type": "application/json; charset=UTF-8"},
	}

	var slackResponseBody = NewEphemeral()

	if verb != http.MethodPost {
		resp.StatusCode = http.StatusForbidden
		slackResponseBody.Text = "Method invalid"
		b, _ := json.Marshal(slackResponseBody)
		resp.Body = string(b)
		return resp, errors.New("Forbidden")
	}

	var slackReq = slackRequestFrom(v)

	if !isTokenValid(slackReq.Token) {
		fmt.Printf("Invalid token in request: %v from postForm: %v", slackReq, v)
		resp.StatusCode = http.StatusUnauthorized
		slackResponseBody.Text = "Unauthorised"
		b, _ := json.Marshal(slackResponseBody)
		resp.Body = string(b)
		return resp, errors.New("Unauthorised")
	}

	if len(slackReq.Text) == 0 {
		sr := NewEphemeral()
		sr.Text = fmt.Sprint("This slack integration provides four options:\n\n-`/tube status` or `/tube status <lineName>` for example `/tube status bakerloo`\n-`/tube subscribe <lineName>`, for example `/tube subscribe bakerloo`\n-`/tube for <username>`, for example `/tube for @jlennon`\n-`/tube version`\n\nFor more details please visit <https://thoeni.io/project/slack-tube-service/|slack-tube-service project page>")
		resp.StatusCode = http.StatusOK
		b, _ := json.Marshal(sr)
		resp.Body = string(b)
		return resp, nil
	}
	slackInput := strings.Split(slackReq.Text[0], " ")

	slackCommand := slackInput[0]
	slackCommandArgs := slackInput[1:]

	switch slackCommand {
	case "status":
		sr, _ := statusCommand(l.tfl, slackCommandArgs, slackReq)
		resp.StatusCode = http.StatusOK
		b, _ := json.Marshal(sr)
		resp.Body = string(b)
	case "for":
		sr, _ := forCommand(l.tfl, l.linesRepo, slackCommandArgs, slackReq)
		resp.StatusCode = http.StatusOK
		b, _ := json.Marshal(sr)
		resp.Body = string(b)
	case "subscribe":
		sr, _ := subscribeCommand(l.tfl, l.userRepo, slackCommandArgs, slackReq)
		resp.StatusCode = http.StatusOK
		b, _ := json.Marshal(sr)
		resp.Body = string(b)
	case "version":
		sr, _ := versionCommand(slackCommandArgs, slackReq)
		resp.StatusCode = http.StatusOK
		b, _ := json.Marshal(sr)
		resp.Body = string(b)
	default:
		sr := NewEphemeral()
		sr.Text = fmt.Sprintf("Unrecognised command: %s", slackCommand)
		resp.StatusCode = http.StatusOK
		b, _ := json.Marshal(sr)
		resp.Body = string(b)
	}

	return resp, nil
}

func slackTokenRequestHandler(verb string, token string, v url.Values) (events.APIGatewayProxyResponse, error) {
	switch verb {
	case http.MethodPut:
		err := validateToken(token)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, err
		}
		tokenStore.AddToken(token)
	case http.MethodDelete:
		tokenStore.DeleteToken(token)
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusAccepted}, nil
}

func slackRequestFrom(v url.Values) slackRequest {
	return slackRequest{
		Token:       v.Get("token"),
		TeamID:      v.Get("teamID"),
		TeamDomain:  v.Get("teamDomain"),
		ChannelID:   v.Get("channelID"),
		ChannelName: v.Get("channelName"),
		UserID:      v.Get("userID"),
		Username:    v.Get("username"),
		Command:     v.Get("command"),
		Text:        []string{v.Get("text")},
		ResponseURL: v.Get("responseURL"),
	}
}
