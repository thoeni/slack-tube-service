package main

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestDecodeTflResponse(t *testing.T) {
	path := "tflResponse"
	inFile, _ := os.OpenFile(path, os.O_RDONLY, 0666)

	statuses, err := decodeTflResponse(inFile)
	if err != nil {
		t.Error("File could not be unmarshalled into a status array")
	}

	if len(statuses) != 11 {
		t.Error("Unmarshalled the incorrect number of statuses.")
	}
}
