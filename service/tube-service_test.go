package service

import (
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/thoeni/go-tfl"
	"github.com/thoeni/slack-tube-service/mocks"
	"testing"
	"time"
)

var validTflClientResponse = []tfl.Report{
	{"line1", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}},
	{"line2", []tfl.Status{{StatusSeverity: 1, Reason: "", StatusSeverityDescription: ""}}},
	{"line3", []tfl.Status{{StatusSeverity: 2, Reason: "", StatusSeverityDescription: ""}}},
}

func TestGetStatusFor_whenValuesReturnedByTflClient(t *testing.T) {
	s, c := initialiseServiceWithTflClientReponse(t, validTflClientResponse, nil)
	defer c.Finish()

	result, _ := s.GetStatusFor([]string{"line1", "line3"})

	if !(len(result) == 2) {
		t.Errorf("Failed to retrieve two lines. There were %d lines.", len(result))
	}
	actualSeverity := result["line3"].LineStatuses[0].StatusSeverity
	expectedSeverity := 2
	if actualSeverity != expectedSeverity {
		t.Errorf("Status severity for line3 in test was %d instead of %d", actualSeverity, expectedSeverity)
	}
}

func TestGetStatusFor_whenEmptyValuesReturnedByTflClient(t *testing.T) {
	s, c := initialiseServiceWithTflClientReponse(t, []tfl.Report{}, nil)
	defer c.Finish()

	result, _ := s.GetStatusFor([]string{"line1", "line3"})

	if !(len(result) == 0) {
		t.Errorf("There should be no lines in the result. There are %d instead", len(result))
	}
}

func TestGetStatusFor_whenNoLinesSpecified_thenAllLinesAreReturned(t *testing.T) {
	s, c := initialiseServiceWithTflClientReponse(t, validTflClientResponse, nil)
	defer c.Finish()

	result, _ := s.GetStatusFor([]string{})

	if !(len(result) == 3) {
		t.Errorf("All lines should be in the result as no lines were specified. There were %d instead.", len(result))
	}
}

func TestGetStatusFor_whenTflClientReturnsError(t *testing.T) {
	s, c := initialiseServiceWithTflClientReponse(t, validTflClientResponse, errors.New("Something went wrong"))
	defer c.Finish()

	result, e := s.GetStatusFor([]string{})

	if e == nil {
		t.Error("Service should propagate Tfl error")
	}

	if !(len(result) == 0) {
		t.Errorf("There should be no lines in the result. There are %d instead", len(result))
	}
}

func TestInMemoryCachedClient_WhenTimeLessThanInvalidate_ThenDoesNotCallTflClientAgain(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockTflClient := mocks.NewMockClient(mockCtrl)
	mockTflClient.EXPECT().GetTubeStatus().Times(0)
	c := InMemoryCachedClient{
		client: mockTflClient,
		invalidateIntervalInSeconds: 10,
		lastRetrieve:                time.Now().Add(-5 * time.Second),
		tubeStatus:                  validTflClientResponse,
	}
	c.GetTubeStatus()
}

func TestInMemoryCachedClient_WhenTimeLessThanInvalidate_ThenDoesCallTflClientToGetFreshData(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockTflClient := mocks.NewMockClient(mockCtrl)
	mockTflClient.EXPECT().GetTubeStatus().Times(1)
	c := InMemoryCachedClient{
		client: mockTflClient,
		invalidateIntervalInSeconds: 10,
		lastRetrieve:                time.Now().Add(-15 * time.Second),
		tubeStatus:                  validTflClientResponse,
	}
	c.GetTubeStatus()
}

func TestFilter(t *testing.T) {
	onlyOneLine := filter(tfl.ReportArrayToMap(validTflClientResponse), []string{"line2"})
	actualLength := len(onlyOneLine)
	if actualLength != 1 {
		t.Errorf("Actual length was %d instead of expected 1", actualLength)
	}
}

func initialiseServiceWithTflClientReponse(t *testing.T, r []tfl.Report, e error) (TubeService, *gomock.Controller) {
	mockCtrl := gomock.NewController(t)
	mockTflClient := mocks.NewMockClient(mockCtrl)
	mockTflClient.EXPECT().GetTubeStatus().Return(r, e)
	return TubeService{client: mockTflClient}, mockCtrl
}

func initialiseInMemoryCachedClientWithTflClientReponse(t *testing.T, r []tfl.Report) (InMemoryCachedClient, *gomock.Controller) {
	mockCtrl := gomock.NewController(t)
	mockTflClient := mocks.NewMockClient(mockCtrl)
	mockTflClient.EXPECT().GetTubeStatus().Return(r, nil)
	return InMemoryCachedClient{
		client: mockTflClient,
		invalidateIntervalInSeconds: 1,
	}, mockCtrl
}
