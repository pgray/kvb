package db

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

func ce(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Page struct {
	Title string
	Body  []byte
}

func Sections(DB *bolt.DB) []string {
	var buckets []string
	err := DB.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			buckets = append(buckets, string(name))
			return nil
		})
	})
	ce(err)
	return buckets
}

func Posts(DB *bolt.DB, section string) []string {
	if section == "" {
		return []string{""}
	}

	var keys []string
	err := DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(section))
		if bucket == nil {
			return fmt.Errorf("No such bucket")
		}
		bucket.ForEach(func(name []byte, _ []byte) error {
			keys = append(keys, string(name))
			return nil
		})
		return nil
	})
	ce(err)
	return keys
}

//Sections directly correlate to buckets
func LoadPage(DB *bolt.DB, section string, title string) Page {
	var body []byte
	err := DB.View(func(tx *bolt.Tx) error {
		bucket := []byte(section)
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		body = append(body, b.Get([]byte(title))...)
		return nil
	})
	ce(err)
	return Page{Title: title, Body: body}
}

//Pages are saved in a SECTION (bucket) by their TITLE (key) in a BODY (the key's value)
func SavePage(DB *bolt.DB, section string, page Page) error {
	err := DB.Update(func(tx *bolt.Tx) error {
		bucket := []byte(section)
		b, err := tx.CreateBucketIfNotExists(bucket)
		ce(err)
		err = b.Put([]byte(page.Title), page.Body)
		ce(err)
		return nil
	})
	ce(err)
	return nil
}
