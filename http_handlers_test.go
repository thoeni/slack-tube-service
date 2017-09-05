package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/thoeni/go-tfl"
	"github.com/thoeni/slack-tube-service/mocks"
)

const (
	tflResponseJson string = "test-data/tflResponse.json"
	bakerloo        string = "\"Bakerloo\":{\"Name\":\"Bakerloo\",\"LineStatuses\":[{\"StatusSeverity\":5,\"StatusSeverityDescription\":\"\",\"Reason\":\"\"}]}"
	jubilee         string = "\"Jubilee\":{\"Name\":\"Jubilee\",\"LineStatuses\":[{\"StatusSeverity\":5,\"StatusSeverityDescription\":\"\",\"Reason\":\"\"}]}"
	waterlooAndCity string = "\"Waterloo & City\":{\"Name\":\"Waterloo \u0026 City\",\"LineStatuses\":[{\"StatusSeverity\":5,\"StatusSeverityDescription\":\"\",\"Reason\":\"\"}]}"
)

func TestLineStatusHandler_whenCalledToRetrieveAllLines(t *testing.T) {

	allLinesTflServiceResponse := tubeServiceResponseGeneratorFor([]string{"Bakerloo", "Jubilee", "Waterloo & City"})
	expectedBody := fmt.Sprintf("{%s,%s,%s}\n", bakerloo, jubilee, waterlooAndCity)

	var c *gomock.Controller = gomock.NewController(t)
	var forAllLines []string
	tubeService = newMockTflService(c, forAllLines, allLinesTflServiceResponse, nil)
	responseRecorder := httptest.NewRecorder()
	defer c.Finish()
	var req *http.Request = httptest.NewRequest(http.MethodGet, "/api/tubestatus/", nil)

	newRouter().ServeHTTP(responseRecorder, req)

	resp := responseRecorder.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, []byte(expectedBody), body)
}

func TestLineStatusHandler_whenCalledToRetrieveSingleLine(t *testing.T) {

	singleLineTflServiceResponse := tubeServiceResponseGeneratorFor([]string{"Bakerloo"})
	expectedBody := fmt.Sprintf("{%s}\n", bakerloo)

	var c *gomock.Controller = gomock.NewController(t)
	forSingleLine := []string{"Bakerloo"}
	tubeService = newMockTflService(c, forSingleLine, singleLineTflServiceResponse, nil)
	responseRecorder := httptest.NewRecorder()
	defer c.Finish()
	var req *http.Request = httptest.NewRequest(http.MethodGet, "/api/tubestatus/Bakerloo", nil)

	newRouter().ServeHTTP(responseRecorder, req)

	resp := responseRecorder.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, []byte(expectedBody), body)
}

func TestLineStatusHandler_whenServiceFails_Returns500(t *testing.T) {

	singleLineTflServiceResponse := tubeServiceResponseGeneratorFor([]string{"Bakerloo"})

	var c *gomock.Controller = gomock.NewController(t)
	forSingleLine := []string{"Bakerloo"}
	tubeService = newMockTflService(c, forSingleLine, singleLineTflServiceResponse, errors.New("Something went wrong"))
	responseRecorder := httptest.NewRecorder()
	defer c.Finish()
	var req *http.Request = httptest.NewRequest(http.MethodGet, "/api/tubestatus/Bakerloo", nil)

	newRouter().ServeHTTP(responseRecorder, req)

	resp := responseRecorder.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, 500, resp.StatusCode)
	assert.Equal(t, []byte("\"There was an error getting information from TFL\"\n"), body)
}

func TestLineStatusHandler_whenServiceReturnsEmptyLinesAndNoError_Returns404(t *testing.T) {

	var noLineTflServiceResponse map[string]tfl.Report

	var c *gomock.Controller = gomock.NewController(t)
	forUnknownLine := []string{"unknownLine"}
	tubeService = newMockTflService(c, forUnknownLine, noLineTflServiceResponse, nil)
	defer c.Finish()

	responseRecorder := httptest.NewRecorder()
	var req *http.Request = httptest.NewRequest(http.MethodGet, "/api/tubestatus/unknownLine", nil)

	newRouter().ServeHTTP(responseRecorder, req)

	resp := responseRecorder.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, 404, resp.StatusCode)
	assert.Equal(t, []byte("\"Line requested not found\"\n"), body)
}

func TestLineStatusHandler_Integration_HappyPathAllLines(t *testing.T) {
	mockTflResponse, _ := ioutil.ReadFile(tflResponseJson)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, string(mockTflResponse))
	}))
	defer ts.Close()

	cachedTflClient := tfl.NewCachedClient(http.DefaultClient, 120)
	cachedTflClient.SetBaseURL(ts.URL + "/")
	tubeService = TubeService{cachedTflClient}

	responseRecorder := httptest.NewRecorder()
	expectedBodyFromTflResponse := "{\"bakerloo\":{\"Name\":\"Bakerloo\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"central\":{\"Name\":\"Central\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"circle\":{\"Name\":\"Circle\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"district\":{\"Name\":\"District\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"hammersmith & city\":{\"Name\":\"Hammersmith & City\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"jubilee\":{\"Name\":\"Jubilee\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"metropolitan\":{\"Name\":\"Metropolitan\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"northern\":{\"Name\":\"Northern\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"piccadilly\":{\"Name\":\"Piccadilly\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"victoria\":{\"Name\":\"Victoria\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]},\"waterloo & city\":{\"Name\":\"Waterloo & City\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]}}\n"
	var req *http.Request = httptest.NewRequest(http.MethodGet, "/api/tubestatus/", nil)

	newRouter().ServeHTTP(responseRecorder, req)

	resp := responseRecorder.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, []byte(expectedBodyFromTflResponse), body)
}

func TestLineStatusHandler_Integration_HappyPathSingleLine(t *testing.T) {
	mockTflResponse, _ := ioutil.ReadFile(tflResponseJson)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, string(mockTflResponse))
	}))
	defer ts.Close()

	client := tfl.NewClient(http.DefaultClient)
	client.SetBaseURL(ts.URL + "/")

	cachedTflClient := tfl.NewCachedClient(http.DefaultClient, 120)
	cachedTflClient.SetBaseURL(ts.URL + "/")
	tubeService = TubeService{cachedTflClient}

	responseRecorder := httptest.NewRecorder()
	expectedBodyFromTflResponse := "{\"Bakerloo\":{\"Name\":\"Bakerloo\",\"LineStatuses\":[{\"StatusSeverity\":10,\"StatusSeverityDescription\":\"Good Service\",\"Reason\":\"\"}]}}\n"
	var req *http.Request = httptest.NewRequest(http.MethodGet, "/api/tubestatus/Bakerloo", nil)

	newRouter().ServeHTTP(responseRecorder, req)

	resp := responseRecorder.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, []byte(expectedBodyFromTflResponse), body)
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

func tubeServiceResponseGeneratorFor(lines []string) map[string]tfl.Report {

	linesMap := make(map[string]tfl.Report)
	linesMap["Waterloo & City"] = tfl.Report{"Waterloo & City", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}}
	linesMap["Bakerloo"] = tfl.Report{"Bakerloo", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}}
	linesMap["Jubilee"] = tfl.Report{"Jubilee", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}}

	var serviceOutputMap = make(map[string]tfl.Report)

	for _, line := range lines {
		serviceOutputMap[line] = linesMap[line]
	}

	return serviceOutputMap
}
