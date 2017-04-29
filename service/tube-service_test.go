package service

import (
	"github.com/golang/mock/gomock"
	"github.com/thoeni/go-tfl"
	"github.com/thoeni/slack-tube-service/mocks"
	"testing"
)

var validTflClientResponse = []tfl.Report{
	{"line1", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}},
	{"line2", []tfl.Status{{StatusSeverity: 1, Reason: "", StatusSeverityDescription: ""}}},
	{"line3", []tfl.Status{{StatusSeverity: 2, Reason: "", StatusSeverityDescription: ""}}},
}

func TestGetStatusFor(t *testing.T) {
	s, c := initialiseServiceWithTflClientReponse(t, validTflClientResponse)
	defer c.Finish()

	result, _ := s.getStatusFor([]string{"line1", "line3"})

	if !(len(result) == 2) {
		t.Errorf("Failed to retrieve two lines. There were %d lines.", len(result))
	}
	actualSeverity := result["line3"].LineStatuses[0].StatusSeverity
	expectedSeverity := 2
	if actualSeverity != expectedSeverity {
		t.Errorf("Status severity for line3 in test was %d instead of %d", actualSeverity, expectedSeverity)
	}
}

func initialiseServiceWithTflClientReponse(t *testing.T, r []tfl.Report) (TflService, *gomock.Controller) {
	mockCtrl := gomock.NewController(t)
	mockTflClient := mocks.NewMockClient(mockCtrl)
	mockTflClient.EXPECT().GetTubeStatus().Return(validTflClientResponse, nil)
	return HttpTubeService{client: mockTflClient}, mockCtrl
}
