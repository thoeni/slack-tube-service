package main

import (
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
	"regexp"
)

var authorisedTokenSet []string

type TokenStorer interface {
	reloadAuthorisedTokens()
	addSlackToken(token string)
	deleteSlackToken(token string)
	close()
}

type BoltTokenStore struct {
	boltDB *bolt.DB
}
type FileTokenStore struct {
	file *os.File
}

func (b BoltTokenStore) reloadAuthorisedTokens() {

	authorisedTokenSet = make([]string, 0)

	b.boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("token"))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			authorisedTokenSet = append(authorisedTokenSet, string(k[:]))
		}

		return nil
	})
}

func (b BoltTokenStore) addSlackToken(token string) {

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

func (b BoltTokenStore) deleteSlackToken(token string) {

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

func (b BoltTokenStore) close() {
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
	for authTokenId := range authorisedTokenSet {
		if token == authorisedTokenSet[authTokenId] {
			return true
		}
	}
	return false
}
