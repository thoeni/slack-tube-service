package main
//
//import (
//	"bytes"
//	"fmt"
//	"github.com/golang/mock/gomock"
//	"github.com/stretchr/testify/assert"
//	"github.com/thoeni/go-tfl"
//	"github.com/thoeni/slack-tube-service/mocks"
//	"io/ioutil"
//	"net/http"
//	"net/http/httptest"
//	"net/url"
//	"strconv"
//	"testing"
//	"github.com/thoeni/slack-tube-service/tflondon"
//)
//
//const (
//	bakerlooSlack        string = "{\"fallback\":\"\",\"color\":\"#B36305\",\"pretext\":\"\",\"text\":\":rage:  *Bakerloo* :: \\n\",\"mrkdwn_in\":[\"text\"]}"
//	jubileeSlack         string = "{\"fallback\":\"\",\"color\":\"#A0A5A9\",\"pretext\":\"\",\"text\":\":rage:  *Jubilee* :: \\n\",\"mrkdwn_in\":[\"text\"]}"
//	waterlooAndCitySlack string = "{\"fallback\":\"\",\"color\":\"#95CDBA\",\"pretext\":\"\",\"text\":\":rage:  *Waterloo & City* :: \\n\",\"mrkdwn_in\":[\"text\"]}"
//)
//
//func TestReportMapToSortedAttachmentsArray_whenInputMap_thenOutputArrayIsSorted(t *testing.T) {
//
//	inputMap := make(map[string]tfl.Report, 3)
//	inputMap["Waterloo & City"] = tfl.Report{"Waterloo & City", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}}
//	inputMap["Bakerloo"] = tfl.Report{"Bakerloo", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}}
//	inputMap["Jubilee"] = tfl.Report{"Jubilee", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}}
//
//	outputArray := reportMapToSortedAttachmentsArray(inputMap)
//
//	assert.Contains(t, outputArray[0].Text, "Bakerloo")
//	assert.Contains(t, outputArray[1].Text, "Jubilee")
//	assert.Contains(t, outputArray[2].Text, "Waterloo & City")
//}
//
//func TestSlackStatusHandler_whenCalledToRetrieveAllLines(t *testing.T) {
//
//	allLinesTflServiceResponse := tubeServiceResponseGeneratorFor([]string{"Bakerloo", "Jubilee", "Waterloo & City"})
//	expectedBody := fmt.Sprintf("{\"text\":\"Slack Tube Service\",\"response_type\":\"ephemeral\",\"attachments\":[%s,%s,%s]}\n", bakerlooSlack, jubileeSlack, waterlooAndCitySlack)
//
//	var c *gomock.Controller = gomock.NewController(t)
//	var forAllLines []string
//	tokenStore = newMockTokenStore(c, []string{"validToken123"}, nil)
//	tubeService = newMockTflService(c, forAllLines, allLinesTflServiceResponse, nil)
//	responseRecorder := httptest.NewRecorder()
//	defer c.Finish()
//
//	data := url.Values{}
//	data.Set("token", "validToken123")
//	data.Set("text", "status")
//	var req *http.Request = httptest.NewRequest(http.MethodPost, "/api/slack/tubestatus/", bytes.NewBufferString(data.Encode()))
//	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
//	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
//
//	_, authorisedTokenSet = tokenStore.RetrieveAllTokens()
//	newRouter().ServeHTTP(responseRecorder, req)
//
//	resp := responseRecorder.Result()
//	body, _ := ioutil.ReadAll(resp.Body)
//
//	assert.Equal(t, 200, resp.StatusCode)
//	assert.Equal(t, []byte(expectedBody), body)
//}
//
//func TestSlackStatusHandler_whenCalledToRetrieveSingleLine(t *testing.T) {
//
//	allLinesTflServiceResponse := tubeServiceResponseGeneratorFor([]string{"Bakerloo"})
//	expectedBody := fmt.Sprintf("{\"text\":\"Slack Tube Service\",\"response_type\":\"ephemeral\",\"attachments\":[%s]}\n", bakerlooSlack)
//
//	var c *gomock.Controller = gomock.NewController(t)
//	var forBakerlooLine []string = []string{"bakerloo"}
//	tokenStore = newMockTokenStore(c, []string{"validToken123"}, nil)
//	tubeService = newMockTflService(c, forBakerlooLine, allLinesTflServiceResponse, nil)
//	responseRecorder := httptest.NewRecorder()
//	defer c.Finish()
//
//	data := url.Values{}
//	data.Set("token", "validToken123")
//	data.Add("text", "status bakerloo")
//	var req *http.Request = httptest.NewRequest(http.MethodPost, "/api/slack/tubestatus/", bytes.NewBufferString(data.Encode()))
//	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
//	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
//
//	_, authorisedTokenSet = tokenStore.RetrieveAllTokens()
//	newRouter().ServeHTTP(responseRecorder, req)
//
//	resp := responseRecorder.Result()
//	body, _ := ioutil.ReadAll(resp.Body)
//
//	assert.Equal(t, 200, resp.StatusCode)
//	assert.Equal(t, []byte(expectedBody), body)
//}
//
//func TestSlackStatusHandler_whenCalledToRetrieveUnexistingLine_thenReturnNotFound(t *testing.T) {
//
//	var noLines map[string]tfl.Report
//
//	var c *gomock.Controller = gomock.NewController(t)
//	var forUnknownLine []string = []string{"unknownLine"}
//	tokenStore = newMockTokenStore(c, []string{"validToken123"}, nil)
//	tubeService = newMockTflService(c, forUnknownLine, noLines, nil)
//	responseRecorder := httptest.NewRecorder()
//	defer c.Finish()
//
//	data := url.Values{}
//	data.Set("token", "validToken123")
//	data.Add("text", "status unknownLine")
//	var req *http.Request = httptest.NewRequest(http.MethodPost, "/api/slack/tubestatus/", bytes.NewBufferString(data.Encode()))
//	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
//	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
//
//	_, authorisedTokenSet = tokenStore.RetrieveAllTokens()
//	newRouter().ServeHTTP(responseRecorder, req)
//
//	resp := responseRecorder.Result()
//	body, _ := ioutil.ReadAll(resp.Body)
//
//	assert.Equal(t, 200, resp.StatusCode)
//	assert.Contains(t, string(body), "\"text\":\"Not a recognised line.\"")
//}
//
//func TestSlackStatusHandler_whenMissingToken_thenReturnUnauthorised(t *testing.T) {
//
//	var c *gomock.Controller = gomock.NewController(t)
//	tokenStore = newMockTokenStore(c, []string{"validToken123"}, nil)
//	responseRecorder := httptest.NewRecorder()
//	defer c.Finish()
//
//	data := url.Values{}
//	data.Set("token", "")
//	data.Add("text", "bakerloo")
//	var req *http.Request = httptest.NewRequest(http.MethodPost, "/api/slack/tubestatus/", bytes.NewBufferString(data.Encode()))
//	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
//	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
//
//	_, authorisedTokenSet = tokenStore.RetrieveAllTokens()
//	newRouter().ServeHTTP(responseRecorder, req)
//
//	resp := responseRecorder.Result()
//
//	assert.Equal(t, 401, resp.StatusCode)
//}
//
//func TestSlackStatusHandler_whenRequestInvalid_thenReturnBadRequest(t *testing.T) {
//
//	var c *gomock.Controller = gomock.NewController(t)
//	tokenStore = newMockTokenStore(c, []string{"validToken123"}, nil)
//	responseRecorder := httptest.NewRecorder()
//	defer c.Finish()
//
//	data := url.Values{}
//	data.Set("token", "validToken123")
//	data.Add("textInvalid", "bakerloo")
//	var req *http.Request = httptest.NewRequest(http.MethodPost, "/api/slack/tubestatus/", bytes.NewBufferString(data.Encode()))
//	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
//	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
//
//	_, authorisedTokenSet = tokenStore.RetrieveAllTokens()
//	newRouter().ServeHTTP(responseRecorder, req)
//
//	resp := responseRecorder.Result()
//
//	assert.Equal(t, 400, resp.StatusCode)
//}
//
//func newMockTokenStore(c *gomock.Controller, output []string, e error) TokenRepository {
//	mockTokenStore := mocks.NewMockRepository(c)
//	mockTokenStore.EXPECT().RetrieveAllTokens().Return(e, output)
//	return mockTokenStore
//}
//
//
//type responseWriterMock struct {
//	t                  *testing.T
//	expectedStatusCode int
//	expectedBody       string
//}
//
//func (responseWriterMock) Header() http.Header {
//	return make(map[string][]string)
//}
//func (m responseWriterMock) Write(b []byte) (int, error) {
//	comparison := bytes.Compare(b, []byte(m.expectedBody))
//	if comparison != 0 {
//		m.t.Errorf("Content of body was:\n\n%s\ninstead of\n\n%s", string(b), m.expectedBody)
//	}
//	return 0, nil
//}
//func (m responseWriterMock) WriteHeader(s int) {
//	if s != m.expectedStatusCode {
//		m.t.Errorf("Status code in header was %d instead of expected: %d", s, m.expectedStatusCode)
//	}
//}
//
//func newMockTflService(c *gomock.Controller, input []string, output map[string]tfl.Report, e error) tflondon.Service {
//	mockTflService := mocks.NewMockTflService(c)
//	mockTflService.EXPECT().GetStatusFor(input).Return(output, e)
//	return mockTflService
//}
//
//func tubeServiceResponseGeneratorFor(lines []string) map[string]tfl.Report {
//
//	linesMap := make(map[string]tfl.Report)
//	linesMap["Waterloo & City"] = tfl.Report{"Waterloo & City", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}}
//	linesMap["Bakerloo"] = tfl.Report{"Bakerloo", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}}
//	linesMap["Jubilee"] = tfl.Report{"Jubilee", []tfl.Status{{StatusSeverity: 5, Reason: "", StatusSeverityDescription: ""}}}
//
//	var serviceOutputMap = make(map[string]tfl.Report)
//
//	for _, line := range lines {
//		serviceOutputMap[line] = linesMap[line]
//	}
//
//	return serviceOutputMap
//}
