package main

import (
	"github.com/thoeni/go-tfl"
	"strings"
	"testing"
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
