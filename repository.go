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

type Repository interface {
	RetrieveAllTokens() (error, []string)
	AddToken(token string)
	DeleteToken(token string)
	Close()
}

type boltRepository struct {
	boltDB *bolt.DB
}
type fileRepository struct {
	file *os.File
}

func (b boltRepository) Close() {
	b.boltDB.Close()
}

// Token

func (b boltRepository) RetrieveAllTokens() (error, []string) {

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

func (b boltRepository) AddToken(token string) {

	fmt.Println("Received request to add new token:", token)

	if err := b.boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("token"))
		err := b.Put([]byte(token), []byte("enabled"))
		return err
	}); err != nil {
		log.Print("Transaction rolled back -> ", err)
	}
	b.RetrieveAllTokens()
}

func (b boltRepository) DeleteToken(token string) {

	fmt.Println("Received request to delete token:", token)

	if err := b.boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("token"))
		err := b.Delete([]byte(token))
		return err
	}); err != nil {
		log.Print("Transaction rolled back -> ", err)
	}
	_, authorisedTokenSet = b.RetrieveAllTokens()
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
