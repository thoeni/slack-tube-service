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

type repository interface {
	retrieveAllTokens() (error, []string)
	addToken(token string)
	deleteToken(token string)
	close()
}

type boltRepository struct {
	boltDB *bolt.DB
}
type fileRepository struct {
	file *os.File
}

func (b boltRepository) close() {
	b.boltDB.Close()
}

// Token

func (b boltRepository) retrieveAllTokens() (error, []string) {

	tokenSet := make([]string, 0)
	var err error

	if err = b.boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("token"))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			tokenSet = append(tokenSet, string(k[:]))
		}

		return nil
	}); err != nil {
		log.Print("Transaction rolled back -> ", err)
	}

	return err, tokenSet
}

func (b boltRepository) addToken(token string) {

	fmt.Println("Received request to add new token:", token)

	if err := b.boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("token"))
		err := b.Put([]byte(token), []byte("enabled"))
		return err
	}); err != nil {
		log.Print("Transaction rolled back -> ", err)
	}
	b.retrieveAllTokens()
}

func (b boltRepository) deleteToken(token string) {

	fmt.Println("Received request to delete token:", token)

	if err := b.boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("token"))
		err := b.Delete([]byte(token))
		return err
	}); err != nil {
		log.Print("Transaction rolled back -> ", err)
	}
	_, authorisedTokenSet = b.retrieveAllTokens()
}

func validateToken(token string) error {
	var validTokenShould = regexp.MustCompile(`\W`)
	if containsNonWordChar := validTokenShould.MatchString(token); containsNonWordChar {
		return errors.New("Token invalid")
	}
	return nil
}

func isTokenValid(token string) bool {
	for authTokenID := range authorisedTokenSet {
		if token == authorisedTokenSet[authTokenID] {
			return true
		}
	}
	return false
}