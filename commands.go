package main

import (
	"fmt"
	"github.com/pkg/errors"
	tfllib "github.com/thoeni/go-tfl"
	"sort"
	"strings"
	"github.com/thoeni/slack-tube-service/tflondon"
	"github.com/thoeni/slack-tube-service/lines"
	"github.com/thoeni/slack-tube-service/users"
)

func statusCommand(tubeService tflondon.Service, slackCommandArgs []string, slackRequest slackRequest) (*slackResponse, error) {

	var r = NewEphemeral()

	tubeLine := strings.Join(slackCommandArgs, " ")

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
		return &r, errors.New("LineNotRecognised")
	}

	r.Attachments = reportMapToSortedAttachmentsArray(reportsMap)
	return &r, nil
}

// Returns the status for lines a specific user subscribed to
func forCommand(tubeService tflondon.Service, linesRepo lines.Repo, slackCommandArgs []string, slackRequest slackRequest) (*slackResponse, error) {

	var r = NewEphemeral()

	if len(slackCommandArgs) == 0 {
		r.Text = "User name must be provided (e.g. `@john`)"
		return &r, errors.New("no user specified")
	}

	var user = slackCommandArgs[0]
	if strings.ToLower(user) == "me" {
		user = fmt.Sprintf("@%s", slackRequest.Username)
	}

	id := fmt.Sprintf("%s-%s", slackRequest.TeamID, user[1:])

	lines, err := linesRepo.GetLinesFor(id)
	if err != nil {
		if err.Error() == "UserNotFound" {
			r.Text = fmt.Sprintf("Couldn't find lines for user: %s", id)
		} else {
			r.Text = fmt.Sprintf("Error while retrieving lines for user: %s", user)
		}
		return &r, errors.Wrap(err, "GetUserError")
	}

	reportsMap, err := tubeService.GetStatusFor(lines)

	if err != nil {
		r.Text = "Error while retrieving information from TFL"
		return &r, errors.Wrap(err, "TFLError")
	} else if len(reportsMap) == 0 {
		r.Text = "Not a recognised line."
		return &r, errors.New("LineNotRecognised")
	}

	r.Attachments = reportMapToSortedAttachmentsArray(reportsMap)
	return &r, nil
}

func subscribeCommand(tubeService tflondon.Service, usersRepo users.Repo, slackCommandArgs []string, slackRequest slackRequest) (*slackResponse, error) {

	var r = NewEphemeral()

	if len(slackCommandArgs) == 0 {
		r.Text = fmt.Sprintf("A line to subscribe to must be specified :thinking_face:. For example `/tube subscribe bakerloo`")
		return &r, errors.New("SubscriptionNotAvailable")
	}

	id := fmt.Sprintf("%s-%s", slackRequest.TeamID, slackRequest.Username)
	username := slackRequest.Username
	subscribedLines := []string{strings.Join(slackCommandArgs, " ")}

	if _, err := statusCommand(tubeService, slackCommandArgs, slackRequest); err != nil {
		if strings.Contains(err.Error(), "LineNotRecognised") {
			r.Text = fmt.Sprintf("Line %s is not a recognised line, therefore subscription is not available", subscribedLines[0])
			return &r, errors.Wrap(err, "SubscriptionNotAvailable")
		}
	}

	if err := usersRepo.PutNewSlackUser(id, username, subscribedLines); err != nil {
		if strings.Contains(err.Error(), "UserAlreadyExists") {
			if err := usersRepo.UpdateExistingSlackUser(id, username, subscribedLines); err != nil {
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

func reportMapToSortedAttachmentsArray(inputMap map[string]tfllib.Report) []attachment {
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

func versionCommand(slackCommandArgs []string, slackRequest slackRequest) (*slackResponse, error) {
	var r = NewEphemeral()
	r.Text = fmt.Sprintf("Slack Tube Service - %s [%s]", AppVersion, Sha)
	return &r, nil
}
