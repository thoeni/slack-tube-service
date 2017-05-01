package main

import (
	"bytes"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/thoeni/go-tfl"
	"github.com/thoeni/slack-tube-service/mocks"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	tflResponseJson string = "test-data/tflResponse.json"
)

func TestReportMapToSortedAttachmentsArray_whenInputMap_thenOutputArrayIsSorted(t *testing.T) {

	inputMap := make(map[string]tfl.Report, 3)
	inputMap["Waterloo & City"] = tfl.Report{"Waterloo & City", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}}
	inputMap["Bakerloo"] = tfl.Report{"Bakerloo", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}}
	inputMap["Jubilee"] = tfl.Report{"Jubilee", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}}

	outputArray := reportMapToSortedAttachmentsArray(inputMap)

	if !strings.Contains(outputArray[0].Text, "Bakerloo") {
		t.Errorf("The first element contained: %s", outputArray[0].Text)
	}
	if !strings.Contains(outputArray[1].Text, "Jubilee") {
		t.Errorf("The second element contained: %s", outputArray[1].Text)
	}
	if !strings.Contains(outputArray[2].Text, "Waterloo & City") {
		t.Errorf("The third element contained: %s", outputArray[2].Text)
	}
}

func TestLineStatusHandler_whenCalledToRetrieveAllLines(t *testing.T) {

	serviceOutputMap, expectedBody := lineStatusHandlerInOut()

	var c *gomock.Controller = gomock.NewController(t)
	var input []string
	tubeService = newMockTflService(c, input, serviceOutputMap, nil)
	responseWriter := responseWriterMock{t, 200, expectedBody}
	defer c.Finish()

	var request http.Request = http.Request{}
	request.RequestURI = "/api/tubestatus/"

	lineStatusHandler(responseWriter, &request)
}

func TestLineStatusHandler_whenServiceFails_Returns500(t *testing.T) {

	serviceOutputMap, _ := lineStatusHandlerInOut()

	var c *gomock.Controller = gomock.NewController(t)
	var input []string
	tubeService = newMockTflService(c, input, serviceOutputMap, errors.New("Something went wrong"))
	responseWriter := responseWriterMock{t, 500, "\"There was an error getting information from TFL\"\n"}
	defer c.Finish()

	var request http.Request = http.Request{}
	request.RequestURI = "/api/tubestatus/"

	lineStatusHandler(responseWriter, &request)
}

func TestLineStatusHandler_whenServiceReturnsEmptyLinesAndNoError_Returns404(t *testing.T) {

	var serviceOutputMap map[string]tfl.Report

	var c *gomock.Controller = gomock.NewController(t)
	var input []string
	tubeService = newMockTflService(c, input, serviceOutputMap, nil)
	responseWriter := responseWriterMock{t, 404, "\"Line requested not found\"\n"}
	defer c.Finish()

	var request http.Request = http.Request{}
	request.RequestURI = "/api/tubestatus/"

	lineStatusHandler(responseWriter, &request)
}

func TestLineStatusHandler_Integration_HappyPath(t *testing.T) {
	mockTflResponse, _ := ioutil.ReadFile(tflResponseJson)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, string(mockTflResponse))
	}))
	defer ts.Close()

	client := tfl.NewClient()
	client.SetBaseURL(ts.URL + "/")

	tflClient = &InMemoryCachedClient{
		client,
		[]tfl.Report{},
		time.Now().Add(-121 * time.Second),
		float64(120),
	}

	tubeService = TubeService{tflClient}

	var request http.Request = http.Request{}
	request.RequestURI = "/api/tubestatus/Bakerloo"

	expectedBodyFromTflResponse := "{\"bakerloo\":{\"Name\":\"Bakerloo\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"central\":{\"Name\":\"Central\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"circle\":{\"Name\":\"Circle\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"district\":{\"Name\":\"District\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"hammersmith & city\":{\"Name\":\"Hammersmith & City\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"jubilee\":{\"Name\":\"Jubilee\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"metropolitan\":{\"Name\":\"Metropolitan\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"northern\":{\"Name\":\"Northern\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"piccadilly\":{\"Name\":\"Piccadilly\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"victoria\":{\"Name\":\"Victoria\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"waterloo & city\":{\"Name\":\"Waterloo & City\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]}}\n"
	var response http.ResponseWriter = responseWriterMock{t, 200, expectedBodyFromTflResponse}

	lineStatusHandler(response, &request)
}

type responseWriterMock struct {
	t                  *testing.T
	expectedStatusCode int
	expectedBody       string
}

func (responseWriterMock) Header() http.Header {
	return make(map[string][]string)
}
func (m responseWriterMock) Write(b []byte) (int, error) {
	comparison := bytes.Compare(b, []byte(m.expectedBody))
	if comparison != 0 {
		m.t.Errorf("Content of body was:\n\n%s\ninstead of\n\n%s", string(b), m.expectedBody)
	}
	return 0, nil
}
func (m responseWriterMock) WriteHeader(s int) {
	if s != m.expectedStatusCode {
		m.t.Errorf("Status code in header was %d instead of expected: %d", s, m.expectedStatusCode)
	}
}

func newMockTflService(c *gomock.Controller, input []string, output map[string]tfl.Report, e error) TflService {
	mockTflService := mocks.NewMockTflService(c)
	mockTflService.EXPECT().GetStatusFor(input).Return(output, e)
	return mockTflService
}

func lineStatusHandlerInOut() (map[string]tfl.Report, string) {
	var serviceOutputMap = make(map[string]tfl.Report)
	serviceOutputMap["Waterloo & City"] = tfl.Report{"Waterloo & City", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}}
	serviceOutputMap["Bakerloo"] = tfl.Report{"Bakerloo", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}}
	serviceOutputMap["Jubilee"] = tfl.Report{"Jubilee", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}}

	var apiOutputBody string = "{\"Bakerloo\":{\"Name\":\"Bakerloo\",\"LineStatuses\":[{\"StatusSeverity\":5,\"StatusSeverityDescription\":\"\",\"Reason\":\"\"}]},\"Jubilee\":{\"Name\":\"Jubilee\",\"LineStatuses\":[{\"StatusSeverity\":5,\"StatusSeverityDescription\":\"\",\"Reason\":\"\"}]},\"Waterloo \u0026 City\":{\"Name\":\"Waterloo \u0026 City\",\"LineStatuses\":[{\"StatusSeverity\":5,\"StatusSeverityDescription\":\"\",\"Reason\":\"\"}]}}\n"

	return serviceOutputMap, apiOutputBody
}
