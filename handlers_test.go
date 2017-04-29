package main

//
//import (
//	"io/ioutil"
//	"net/http/httptest"
//	"fmt"
//	"testing"
//	"net/http"
//	"github.com/thoeni/go-tfl"
//)
//
//func TestLineStatusHandler(t *testing.T) {
//	mockTflResponse, _ := ioutil.ReadFile(testDataCorrect)
//	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.Header().Set("Content-Type", "application/json")
//		fmt.Fprintln(w, string(mockTflResponse))
//	}))
//	defer ts.Close()
//	client := tfl.NewClient()
//	client.SetBaseURL(ts.URL + "/")
//
//	statuses, err := client.GetTubeStatus()
//
//	if err != nil {
//		t.Error("Client failed to retrieve TFL data from mock server")
//	}
//	if len(statuses) != 11 {
//		t.Error("Client retrieved and unmarshalled an incorrect number of statuses")
//	}
//}
