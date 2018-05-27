package tfl

import (
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, expectedSeverity, actualSeverity, "Status severity for line3 in test was %d instead of %d", actualSeverity, expectedSeverity)
}

func TestGetStatusFor_whenEmptyValuesReturnedByTflClient(t *testing.T) {
	s, c := initialiseServiceWithTflClientReponse(t, []tfl.Report{}, nil)
	defer c.Finish()

	result, _ := s.GetStatusFor([]string{"line1", "line3"})

	assert.Equal(t, 0, len(result), "There should be no lines in the result. There are %d instead", len(result))
}

func TestGetStatusFor_whenNoLinesSpecified_thenAllLinesAreReturned(t *testing.T) {
	s, c := initialiseServiceWithTflClientReponse(t, validTflClientResponse, nil)
	defer c.Finish()

	result, _ := s.GetStatusFor([]string{})

	assert.Equal(t, 3, len(result), "All lines should be in the result as no lines were specified. There were %d instead.", len(result))
}

func TestGetStatusFor_whenTflClientReturnsError(t *testing.T) {
	s, c := initialiseServiceWithTflClientReponse(t, validTflClientResponse, errors.New("Something went wrong"))
	defer c.Finish()

	result, e := s.GetStatusFor([]string{})

	if e == nil {
		t.Error("Service should propagate Tfl error")
	}

	assert.Equal(t, 0, len(result), "There should be no lines in the result. There are %d instead", len(result))
}

func TestInMemoryCachedClient_WhenTimeLessThanInvalidate_ThenDoesNotCallTflClientAgain(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockTflClient := mocks.NewMockClient(mockCtrl)
	mockTflClient.EXPECT().GetTubeStatus().Times(0)
	c := tfl.InMemoryCachedClient{
		Client: mockTflClient,
		InvalidateIntervalInSeconds: 10,
		LastUpdated:                 time.Now().Add(-5 * time.Second),
		TubeStatus:                  validTflClientResponse,
	}
	c.GetTubeStatus()
}

func TestInMemoryCachedClient_WhenTimeLessThanInvalidate_ThenDoesCallTflClientToGetFreshData(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockTflClient := mocks.NewMockClient(mockCtrl)
	mockTflClient.EXPECT().GetTubeStatus().Times(1)
	c := tfl.InMemoryCachedClient{
		Client: mockTflClient,
		InvalidateIntervalInSeconds: 10,
		LastUpdated:                 time.Now().Add(-15 * time.Second),
		TubeStatus:                  validTflClientResponse,
	}
	c.GetTubeStatus()
}

func TestInMemoryCachedClient_WhenSetUrl_ThenChangesTheUnderlyingClientURL(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockTflClient := mocks.NewMockClient(mockCtrl)
	mockTflClient.EXPECT().SetBaseURL("newUrl").Times(1)
	c := tfl.InMemoryCachedClient{
		Client: mockTflClient,
		InvalidateIntervalInSeconds: 10,
		LastUpdated:                 time.Now().Add(-15 * time.Second),
		TubeStatus:                  validTflClientResponse,
	}
	c.SetBaseURL("newUrl")
}

func TestFilter(t *testing.T) {
	onlyOneLine := filter(tfl.ReportArrayToMap(validTflClientResponse), []string{"Line2"})
	actualLength := len(onlyOneLine)

	assert.Equal(t, 1, actualLength, "Actual length was %d instead of expected 1", actualLength)
}

func initialiseServiceWithTflClientReponse(t *testing.T, r []tfl.Report, e error) (service, *gomock.Controller) {
	mockCtrl := gomock.NewController(t)
	mockTflClient := mocks.NewMockClient(mockCtrl)
	mockTflClient.EXPECT().GetTubeStatus().Return(r, e)
	return service{Client: mockTflClient}, mockCtrl
}
