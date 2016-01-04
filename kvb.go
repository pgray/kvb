package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	//	"regexp"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/russross/blackfriday"
)

//Globals
var DB *bolt.DB

//Flags
type Config struct {
	dbfile  string
	wport   int
	bport   int
	backups bool
}

var DBFILE string
var WPORT int
var BPORT int
var BACKUPS bool

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
	var body []byte
	err := DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(section))
		body = append(body, b.Get([]byte(title))...)
		return nil
	})
	ce(err)
	return Page{Title: title, Body: body}
}

//Pages are saved in a SECTION (bucket) by their TITLE (key) in a BODY (the key's value)
func savePage(section string, page Page) error {
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

func saveHandler(w http.ResponseWriter, r *http.Request, section string, title string) {
	body := r.FormValue("body")
	savePage(section, Page{Title: title, Body: []byte(body)})
	fmt.Println("save: ", section, title)
	http.Redirect(w, r, "/b/"+section+"/"+title, http.StatusFound)
}

func editHandler(w http.ResponseWriter, r *http.Request, section string, title string) {
	if title == "" {
		title = "root"
	}

	p := loadPage(section, title)
	t, err := template.ParseFiles("templates/edit.html")
	ce(err)
	fmt.Println("edit: ", section, title)
	asdf := struct {
		Section string
		Title   string
		Body    []byte
	}{
		section,
		p.Title,
		p.Body,
	}
	t.Execute(w, asdf)
}

func browseHandler(w http.ResponseWriter, r *http.Request, section string, title string) {
	if title == "" {
		rootHandler(w, r)
	}

	p := loadPage(section, title)

	if p.Body == nil {
		p.Body = append(p.Body, []byte("Sorry, that page does not exist")...)
	}

	t, err := template.ParseFiles("templates/browse.html")
	ce(err)
	fmt.Println("browse: ", section, title)
	asdf := struct {
		Title string
		Body  template.HTML
	}{
		p.Title,
		template.HTML(blackfriday.MarkdownCommon(p.Body)),
	}
	t.Execute(w, asdf)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/b/root/root", http.StatusFound)
}

func initdb() {
	section := "root"
	title := "root"
	err := DB.Update(func(tx *bolt.Tx) error {
		bucket := []byte(section)
		b, err := tx.CreateBucketIfNotExists(bucket)
		ce(err)
		err = b.Put([]byte(title), []byte("This is the root of the blog"))
		ce(err)
		return nil
	})
	ce(err)
}

func backupHandler(w http.ResponseWriter, r *http.Request) {
	err := DB.View(func(tx *bolt.Tx) error {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="backup.db"`)
		w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))

		_, err := tx.WriteTo(w)
		return err
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Raw Path:", r.URL.Path)
		path := r.URL.Path
		path = strings.TrimSpace(path)

		if strings.HasPrefix(path, "/") {
			path = path[1:]
		}

		if strings.HasSuffix(path, "/") {
			cut_off_last_char_len := len(path) - 1
			path = path[:cut_off_last_char_len]
		}
		fmt.Println("Path: ", path)

		m := strings.Split(r.URL.Path, "/")
		fmt.Println("m: ", m, " m length: ", len(m))
		if m == nil {
			http.NotFound(w, r)
			return
		}
		if len(m) < 4 {
			m = append(m, "root")
			m = append(m, "root")
			m = append(m, "root")
		}
		fmt.Println("m:", m, " m length: ", len(m))
		fmt.Println(m[2], m[3])
		fn(w, r, m[2], m[3])
	}
}

func readVars() {
	flag.StringVar(&DBFILE, "database_file", "bolt-kvb.db", "specify a filename for the database (BoltDB: https://github.com/boltdb/bolt)")
	flag.IntVar(&WPORT, "web_port", 8080, "specify the port to listen on for web connections")
	flag.IntVar(&BPORT, "backup_port", 8090, "specify the port to listen on for backups of the database file")
	flag.BoolVar(&BACKUPS, "backups", false, "specify whether the backup port should be enabled (defaults to false)")
	flag.Parse()
}

func main() {
	finish := make(chan bool)

	readVars()

	db, err := bolt.Open(DBFILE, 0600, nil)
	ce(err)
	DB = db //Global Pointer
	defer db.Close()

	initdb()

	webserver := http.NewServeMux()
	webserver.HandleFunc("/", rootHandler)
	webserver.HandleFunc("/b/", makeHandler(browseHandler))
	webserver.HandleFunc("/e/", makeHandler(editHandler))
	webserver.HandleFunc("/s/", makeHandler(saveHandler))

	backupserver := http.NewServeMux()
	backupserver.HandleFunc("/backup", backupHandler)

	fmt.Println("Using database file: ", DBFILE)

	go func() {
		fmt.Println("Web server starting on port: ", WPORT)

		err := http.ListenAndServe(":"+strconv.Itoa(WPORT), webserver)

		if err != nil {
			log.Fatal("webserver ListenAndServe: ", err)
		}
	}()

	if BACKUPS == true {
		go func() {
			fmt.Println("Backups available at /backup on port: ", BPORT)

			err := http.ListenAndServe(":"+strconv.Itoa(BPORT), backupserver)

			if err != nil {
				log.Fatal("backupserver ListenAndServe: ", err)
			}
		}()
	}

	<-finish
}
