package main

import (
	"bytes"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/thoeni/go-tfl"
	"github.com/thoeni/slack-tube-service/mocks"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

const (
	bakerlooSlack        string = "{\"fallback\":\"\",\"color\":\"#B36305\",\"pretext\":\"\",\"text\":\":rage:  *Bakerloo* :: \\n\",\"mrkdwn_in\":[\"text\"]}"
	jubileeSlack         string = "{\"fallback\":\"\",\"color\":\"#A0A5A9\",\"pretext\":\"\",\"text\":\":rage:  *Jubilee* :: \\n\",\"mrkdwn_in\":[\"text\"]}"
	waterlooAndCitySlack string = "{\"fallback\":\"\",\"color\":\"#95CDBA\",\"pretext\":\"\",\"text\":\":rage:  *Waterloo & City* :: \\n\",\"mrkdwn_in\":[\"text\"]}"
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

func TestSlackStatusHandler_whenCalledToRetrieveAllLines(t *testing.T) {

	allLinesTflServiceResponse := tubeServiceResponseGeneratorFor([]string{"Bakerloo", "Jubilee", "Waterloo & City"})
	expectedBody := fmt.Sprintf("{\"text\":\"Slack Tube Service\",\"response_type\":\"ephemeral\",\"attachments\":[%s,%s,%s]}\n", bakerlooSlack, jubileeSlack, waterlooAndCitySlack)

	var c *gomock.Controller = gomock.NewController(t)
	var forAllLines []string
	tokenStore = newMockTokenStore(c, []string{"validToken123"}, nil)
	tubeService = newMockTflService(c, forAllLines, allLinesTflServiceResponse, nil)
	responseRecorder := httptest.NewRecorder()
	defer c.Finish()

	data := url.Values{}
	data.Set("token", "validToken123")
	var req *http.Request = httptest.NewRequest(http.MethodPost, "/api/slack/tubestatus/", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	_, authorisedTokenSet = tokenStore.RetrieveAllTokens()
	newRouter().ServeHTTP(responseRecorder, req)

	resp := responseRecorder.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Errorf("Status code returned was %d instead of expected 200", resp.StatusCode)
	}
	if bytes.Compare(body, []byte(expectedBody)) != 0 {
		t.Errorf("Body and expected body do not match: received:\n%s \n expected:\n%s", string(body), expectedBody)
	}
}

func TestSlackStatusHandler_whenCalledToRetrieveSingleLine(t *testing.T) {

	allLinesTflServiceResponse := tubeServiceResponseGeneratorFor([]string{"Bakerloo"})
	expectedBody := fmt.Sprintf("{\"text\":\"Slack Tube Service\",\"response_type\":\"ephemeral\",\"attachments\":[%s]}\n", bakerlooSlack)

	var c *gomock.Controller = gomock.NewController(t)
	var forBakerlooLine []string = []string{"bakerloo"}
	tokenStore = newMockTokenStore(c, []string{"validToken123"}, nil)
	tubeService = newMockTflService(c, forBakerlooLine, allLinesTflServiceResponse, nil)
	responseRecorder := httptest.NewRecorder()
	defer c.Finish()

	data := url.Values{}
	data.Set("token", "validToken123")
	data.Add("text", "bakerloo")
	var req *http.Request = httptest.NewRequest(http.MethodPost, "/api/slack/tubestatus/", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	_, authorisedTokenSet = tokenStore.RetrieveAllTokens()
	newRouter().ServeHTTP(responseRecorder, req)

	resp := responseRecorder.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Errorf("Status code returned was %d instead of expected 200", resp.StatusCode)
	}
	if bytes.Compare(body, []byte(expectedBody)) != 0 {
		t.Errorf("Body and expected body do not match: received:\n%s \n expected:\n%s", string(body), expectedBody)
	}
}

func TestSlackStatusHandler_whenCalledToRetrieveUnexistingLine_thenReturnNotFound(t *testing.T) {

	var noLines map[string]tfl.Report

	var c *gomock.Controller = gomock.NewController(t)
	var forUnknownLine []string = []string{"unknownLine"}
	tokenStore = newMockTokenStore(c, []string{"validToken123"}, nil)
	tubeService = newMockTflService(c, forUnknownLine, noLines, nil)
	responseRecorder := httptest.NewRecorder()
	defer c.Finish()

	data := url.Values{}
	data.Set("token", "validToken123")
	data.Add("text", "unknownLine")
	var req *http.Request = httptest.NewRequest(http.MethodPost, "/api/slack/tubestatus/", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	_, authorisedTokenSet = tokenStore.RetrieveAllTokens()
	newRouter().ServeHTTP(responseRecorder, req)

	resp := responseRecorder.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Errorf("Status code returned was %d instead of expected 200", resp.StatusCode)
	}
	if !strings.Contains(string(body), "\"text\":\"Not a recognised line.\"") {
		t.Errorf("Body did not contain the expected substring:\n%s \n expected substring:\n%s", string(body), "\"text\":\"Not a recognised line\".")
	}
}

func TestSlackStatusHandler_whenMissingToken_thenReturnUnauthorised(t *testing.T) {

	var c *gomock.Controller = gomock.NewController(t)
	tokenStore = newMockTokenStore(c, []string{"validToken123"}, nil)
	responseRecorder := httptest.NewRecorder()
	defer c.Finish()

	data := url.Values{}
	data.Set("token", "")
	data.Add("text", "bakerloo")
	var req *http.Request = httptest.NewRequest(http.MethodPost, "/api/slack/tubestatus/", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	_, authorisedTokenSet = tokenStore.RetrieveAllTokens()
	newRouter().ServeHTTP(responseRecorder, req)

	resp := responseRecorder.Result()
	if resp.StatusCode != 401 {
		t.Errorf("Status code returned was %d instead of expected 401", resp.StatusCode)
	}
}

func TestSlackStatusHandler_whenRequestInvalid_thenReturnBadRequest(t *testing.T) {

	var c *gomock.Controller = gomock.NewController(t)
	tokenStore = newMockTokenStore(c, []string{"validToken123"}, nil)
	responseRecorder := httptest.NewRecorder()
	defer c.Finish()

	data := url.Values{}
	data.Set("token", "validToken123")
	data.Add("textInvalid", "bakerloo")
	var req *http.Request = httptest.NewRequest(http.MethodPost, "/api/slack/tubestatus/", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	_, authorisedTokenSet = tokenStore.RetrieveAllTokens()
	newRouter().ServeHTTP(responseRecorder, req)

	resp := responseRecorder.Result()
	if resp.StatusCode != 400 {
		t.Errorf("Status code returned was %d instead of expected 400", resp.StatusCode)
	}
}

func newMockTokenStore(c *gomock.Controller, output []string, e error) Repository {
	mockTokenStore := mocks.NewMockRepository(c)
	mockTokenStore.EXPECT().RetrieveAllTokens().Return(e, output)
	return mockTokenStore
}
