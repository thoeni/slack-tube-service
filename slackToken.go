package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/boltdb/bolt"
)

var authorisedTokenSet []string

type tokenStorer interface {
	reloadAuthorisedTokens()
	addSlackToken(token string)
	deleteSlackToken(token string)
	close()
}

type boltTokenStore struct {
	boltDB *bolt.DB
}
type fileTokenStore struct {
	file *os.File
}

func (b boltTokenStore) reloadAuthorisedTokens() {

	authorisedTokenSet = make([]string, 0)

	if err := b.boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("token"))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			authorisedTokenSet = append(authorisedTokenSet, string(k[:]))
		}

		return nil
	}); err != nil {
		log.Print("Transaction rolled back -> ", err)
	}
}

func (b boltTokenStore) addSlackToken(token string) {

	fmt.Println("Received request to add new token:", token)

	if err := b.boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("token"))
		err := b.Put([]byte(token), []byte("enabled"))
		return err
	}); err != nil {
		log.Print("Transaction rolled back -> ", err)
	}
	b.reloadAuthorisedTokens()
}

func (b boltTokenStore) deleteSlackToken(token string) {

	fmt.Println("Received request to delete token:", token)

	if err := b.boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("token"))
		err := b.Delete([]byte(token))
		return err
	}); err != nil {
		log.Print("Transaction rolled back -> ", err)
	}
	b.reloadAuthorisedTokens()
}

func (b boltTokenStore) close() {
	b.boltDB.Close()
}

func validateToken(token string) error {
	var validTokenShould = regexp.MustCompile(`\W`)
	if containsNonWordChar := validTokenShould.MatchString(token); containsNonWordChar {
		return errors.New("Token invalid")
	}
	return nil
}

func isTokenValid(token string) bool {
	fmt.Printf("Token is: %v. Auth token set is: %v", token, authorisedTokenSet)
	for authTokenID := range authorisedTokenSet {
		if token == authorisedTokenSet[authTokenID] {
			return true
		}
	}
	return false
}
