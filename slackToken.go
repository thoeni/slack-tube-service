package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
)

var authorisedTokenSet []string

const authorisedTokenFileLocation = "./tokens.txt"

func loadAuthorisedTokensFromFile(path string) {
	inFile, _ := os.Open(path)
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)

	authorisedTokenSet = make([]string, 0)

	for scanner.Scan() {
		authorisedTokenSet = append(authorisedTokenSet, scanner.Text())
	}
}

func addSlackToken(token string) {
	fmt.Println("Received request to add new token:", token)
	addAuthorisedTokenToFile(authorisedTokenFileLocation, []string{token})
	loadAuthorisedTokensFromFile(authorisedTokenFileLocation)
}

func deleteSlackToken(token string) {
	fmt.Println("Received request to delete token:", token)
	deleteAuthorisedTokenFromMemory(token)
	go flushTokenSetToFile(authorisedTokenFileLocation)
}

func flushTokenSetToFile(path string) {
	err := os.Remove(path)
	if err != nil {
		log.Fatal(err)
	}

	addAuthorisedTokenToFile(path, authorisedTokenSet)
}

func addAuthorisedTokenToFile(path string, token []string) {
	inFile, _ := os.OpenFile(path, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	writer := bufio.NewWriter(inFile)
	defer inFile.Close()

	for i := range token {
		fmt.Fprintln(writer, token[i])
	}
	writer.Flush()
}

func deleteAuthorisedTokenFromMemory(token string) {
	max := len(authorisedTokenSet)
	for i := 0; i < max; i++ {
		if authorisedTokenSet[i] == token {
			authorisedTokenSet = append(authorisedTokenSet[:i], authorisedTokenSet[i+1:]...)
			max = len(authorisedTokenSet)
			i--
		}
	}
}

func validateToken(token string) error {
	var validTokenShould = regexp.MustCompile(`\W`)
	if containsNonWordChar := validTokenShould.MatchString(token); containsNonWordChar {
		return errors.New("Token invalid")
	}
	return nil
}

func isTokenValid(token string) bool {
	for authTokenId := range authorisedTokenSet {
		if token == authorisedTokenSet[authTokenId] {
			return true
		}
	}
	return false
}
