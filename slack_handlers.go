package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/pkg/errors"
	"github.com/thoeni/go-tfl"
	"net/http"
	"sort"
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

	slackInput := strings.Split(slackReq.Text[0], " ")
	slackCommand := slackInput[0]
	slackCommandArgs := slackInput[1:]

	switch slackCommand {
	case "status":
		sr, _ := statusCommand(slackCommandArgs, slackReq.TeamDomain)
		w.WriteHeader(http.StatusOK)
		encoder.Encode(sr)
	case "subscribe":
		sr, _ := subscribeCommand(slackCommandArgs, slackReq.TeamID, slackReq.Username)
		w.WriteHeader(http.StatusOK)
		encoder.Encode(sr)
	default:
		sr := NewEphemeral()
		sr.Text = fmt.Sprintf("Unrecognised command: %s", slackCommand)
		w.WriteHeader(http.StatusOK)
		encoder.Encode(sr)
	}
}

func statusCommand(slackCommandArgs []string, domain string) (*slackResponse, error) {

	var r slackResponse = NewEphemeral()

	tubeLine := strings.Join(slackCommandArgs, " ")
	teamDomain := strings.Replace(domain, " ", "", -1)

	go func() {
		tubeLineLabel := "all"
		if tubeLine != "" {
			tubeLineLabel = tubeLine
		}
		slackRequestsTotal.WithLabelValues(teamDomain, tubeLineLabel).Inc()
	}()

	r.Text = fmt.Sprintf("Slack Tube Service")

	var lines []string
	if tubeLine != "" {
		lines = []string{tubeLine}
	}

	reportsMap, err := tubeService.GetStatusFor(lines)

	if err != nil {
		r.Text = "Error while retrieving information from TFL"
		return &r, errors.Wrap(err, "TFLError")
	} else if len(reportsMap) == 0 {
		r.Text = "Not a recognised line."
		return &r, errors.Wrap(err, "LineNotRecognised")
	}

	r.Attachments = reportMapToSortedAttachmentsArray(reportsMap)
	return &r, nil
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

type slackUserItem struct {
	ID              string   `dynamodbav:"id""`
	Username        string   `dynamodbav:"username""`
	SubscribedLines []string `dynamodbav:"subscribedLines""`
}

func subscribeCommand(slackCommandArgs []string, teamId string, username string) (*slackResponse, error) {

	var r slackResponse = NewEphemeral()

	id := fmt.Sprintf("%s-%s", teamId, username)
	subscribedLines := []string{strings.Join(slackCommandArgs, " ")}

	if err := putNewSlackUser(id, username, subscribedLines); err != nil {
		if strings.Contains(err.Error(), "UserAlreadyExists") {
			if err := updateExistingSlackUser(id, username, subscribedLines); err != nil {
				r.Text = fmt.Sprintf("Error while updating subscriptions for user %s", username)
				return &r, nil
			}
			r.Text = fmt.Sprintf("Line %s added to subscriptions for existing user %s", subscribedLines[0], username)
			return &r, nil
		} else {
			r.Text = fmt.Sprintf("Error while creating subscriptions for user %s", username)
			return &r, nil
		}
	}
	r.Text = fmt.Sprintf("Line %s added to subscriptions for new user %s", subscribedLines[0], username)
	return &r, nil
}

func putNewSlackUser(id string, username string, subscribedLines []string) error {

	item, err := dynamodbattribute.MarshalMap(slackUserItem{
		ID:              id,
		Username:        username,
		SubscribedLines: subscribedLines,
	})
	if err != nil {
		return errors.Wrap(err, "Something went wrong while marshalling the user")
	}

	ce := "attribute_not_exists(id)"
	rv := "ALL_OLD"

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName:           aws.String("slack-users"),
		ReturnValues:        &rv,
		Item:                item,
		ConditionExpression: &ce,
	})
	if err != nil {
		if ae, ok := err.(awserr.RequestFailure); ok && ae.Code() == "ConditionalCheckFailedException" {
			return errors.Wrap(err, "UserAlreadyExists")
		} else {
			return errors.Wrap(err, "Something went wrong while inserting the user")
		}
	}
	return nil
}

func updateExistingSlackUser(id string, username string, subscribedLines []string) error {
	idAv, _ := dynamodbattribute.Marshal(id)
	usernameAv, _ := dynamodbattribute.Marshal(username)
	subscribedLinesAv, _ := dynamodbattribute.Marshal(subscribedLines)
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{":valSubLines": subscribedLinesAv, ":username": usernameAv}
	ue := "set username = :username, subscribedLines = list_append(subscribedLines, :valSubLines)"

	_, err := svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName:                 aws.String("slack-users"),
		Key:                       map[string]*dynamodb.AttributeValue{"id": idAv},
		UpdateExpression:          &ue,
		ExpressionAttributeValues: expressionAttributeValues,
	})

	if err != nil {
		return errors.Wrap(err, "UpdateFailed")
	}
	return nil
}
