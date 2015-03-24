package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"html/template"
	"log"
	"net/http"
)

var DBFILE = "test.db"

type Page struct {
	Title string
	Body  []byte
}

func ce(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//Sections directly correlate to buckets
func loadPage(section string, title string) Page {
	db, err := bolt.Open(DBFILE, 0600, nil)
	ce(err)
	// FIGURE OUT CORRECT BODY MEMORY ALLOCATION
	var body []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(section))
		body = b.Get([]byte(title))
		fmt.Println("Everything worked to here")
		fmt.Println("Body: ", string(body))
		return nil
	})
	db.Close()
	fmt.Println("loadPage body: ", string(body))
	return Page{Title: title, Body: body}
}

func handler(w http.ResponseWriter, r *http.Request) {
	title := "title"
	//body := "things are ok"
	//fmt.Println(title)
	//fmt.Println(body)
	//p := &Page{Title: title, Body: body}
	p := loadPage("Main", title)
	fmt.Println("Loaded the page", p.Title, string(p.Body))
	fmt.Println(p.Body)
	t, err := template.ParseFiles("templates/browse.html")
	fmt.Println("Parsed the template")
	ce(err)
	t.Execute(w, p)
}

func setDB() {
	db, err := bolt.Open(DBFILE, 0600, nil)
	ce(err)

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := []byte("Main")
		b, err := tx.CreateBucketIfNotExists(bucket)
		ce(err)
		title := []byte("title")
		body := []byte("that was the title")
		err = b.Put(title, body)
		ce(err)
		return nil
	})
	ce(err)

	db.View(func(tx *bolt.Tx) error {
		title := []byte("title")
		bucket := []byte("Main")
		b := tx.Bucket(bucket)
		titleBack := b.Get(title)
		fmt.Println("Stored: ", string(titleBack))
		return nil
	})
	ce(err)

	err = db.View(func(tx *bolt.Tx) error {
		bucket := []byte("Main")
		title := []byte("title")
		b := tx.Bucket(bucket)
		titleBack := b.Get(title)
		fmt.Println("Body: ", string(titleBack))
		return nil
	})
	ce(err)
	db.Close()
}

func main() {
	setDB()
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
