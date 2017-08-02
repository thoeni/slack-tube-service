package main

import "github.com/thoeni/go-tfl"

type slackResponse struct {
	Text         string       `json:"text"`
	ResponseType string       `json:"response_type"`
	Attachments  []attachment `json:"attachments"`
}

func NewEphemeral() slackResponse {
	return slackResponse{
		ResponseType: "ephemeral",
	}
}

type attachment struct {
	Fallback string   `json:"fallback"`
	Color    string   `json:"color"`
	Pretext  string   `json:"pretext"`
	Text     string   `json:"text"`
	MrkdwnIn []string `json:"mrkdwn_in"`
}

func mapTflLineToSlackAttachment(r tfl.Report) attachment {
	var slackAttachment attachment
	slackAttachment.Text = createSlackText(r)
	slackAttachment.Color = mapLineNameToHexColor(r.Name)
	slackAttachment.MrkdwnIn = []string{"text"}
	return slackAttachment
}

func createSlackText(r tfl.Report) string {
	slackText := ""
	slackSeverity := mapTflStatuServerityToSlackSeverity(r.LineStatuses[0].StatusSeverity)
	slackText = slackText + slackSeverity.Emoji
	slackText = slackText + "  *" + r.Name + "*"
	slackText = slackText + " :: " + r.LineStatuses[0].StatusSeverityDescription
	if slackSeverity == danger || slackSeverity == warning {
		slackText = slackText + "\n" + r.LineStatuses[0].Reason
	}
	return slackText
}

func mapLineNameToHexColor(lineName string) string {
	return lineColors[lineName]
}

func mapTflStatuServerityToSlackSeverity(statusSeverity int) slackSeverity {
	return severity[statusSeverity]
}

var danger = slackSeverity{"danger", ":rage:"}
var warning = slackSeverity{"warning", ":warning:"}
var good = slackSeverity{"good", ":grinning:"}

type slackSeverity struct {
	Color string
	Emoji string
}

var severity = map[int]slackSeverity{
	0:  danger,
	1:  danger,
	2:  danger,
	3:  danger,
	4:  danger,
	5:  danger,
	6:  danger,
	7:  warning,
	8:  warning,
	9:  warning,
	10: good,
	11: danger,
	12: warning,
	13: warning,
	14: warning,
	15: danger,
	16: danger,
	17: danger,
	18: good,
	19: good,
	20: warning,
}

var lineColors = map[string]string{
	"Bakerloo":           "#B36305",
	"Central":            "#E32017",
	"Circle":             "#FFD300",
	"District":           "#00782A",
	"Hammersmith & City": "#F3A9BB",
	"Jubilee":            "#A0A5A9",
	"Metropolitan":       "#9B0056",
	"Northern":           "#000000",
	"Piccadilly":         "#003688",
	"Victoria":           "#0098D4",
	"Waterloo & City":    "#95CDBA",
	"London Overground":  "#EE7C0E",
	"DLR":                "#00A4A7",
	"TfL Rail":           "#0019A8",
}
