package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
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

func sections() []string {
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

func posts(section string) []string {
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
func loadPage(section string, title string) Page {
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
	fmt.Println("saveHandler: ", section, title)
	body := r.FormValue("body")
	savePage(section, Page{Title: title, Body: []byte(body)})
	fmt.Println("save: ", section, title)
	http.Redirect(w, r, "/b/"+section+"/"+title, http.StatusFound)
}

func editHandler(w http.ResponseWriter, r *http.Request, section string, title string) {
	fmt.Println("editHandler: ", section, title)
	if title == "" {
		title = "root"
	}

	p := loadPage(section, title)
	t, err := template.ParseFiles("templates/edit.html")
	ce(err)
	tempStruct := struct {
		Section string
		Title   string
		Body    []byte
	}{
		section,
		p.Title,
		p.Body,
	}
	t.Execute(w, tempStruct)
}

func browseHandler(w http.ResponseWriter, r *http.Request, section string, title string) {
	fmt.Println("browseHandler: ", section, title)
	if title == "" {
		t, err := template.ParseFiles("templates/section.html")
		ce(err)
		posts := posts(section)
		fmt.Println("posts: ", posts)
		tempStruct := struct {
			Section string
			Posts   []string
		}{
			section,
			posts,
		}
		t.Execute(w, tempStruct)
		return
	}

	p := loadPage(section, title)

	if p.Body == nil {
		p.Body = append(p.Body, []byte("Sorry, that page does not exist")...)
	}

	t, err := template.ParseFiles("templates/browse.html")
	ce(err)
	fmt.Println("browse: ", section, title)
	tempStruct := struct {
		Section string
		Title   string
		Body    template.HTML
	}{
		section,
		p.Title,
		template.HTML(blackfriday.MarkdownCommon(p.Body)),
	}
	t.Execute(w, tempStruct)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/root.html")
	ce(err)
	test := sections()
	tempStruct := struct {
		Sections []string
	}{
		test,
	}
	t.Execute(w, tempStruct)
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
		path := r.URL.Path
		path = strings.TrimSpace(path)

		if strings.HasPrefix(path, "/") {
			path = path[1:]
		}

		if strings.HasSuffix(path, "/") {
			cut_off_last_char_len := len(path) - 1
			path = path[:cut_off_last_char_len]
		}

		m := strings.Split(r.URL.Path, "/")
		fmt.Println(m)

		if len(m) < 4 {
			fn(w, r, m[2], "")
		} else {
			fn(w, r, m[2], m[3])
		}
	}
}

func init() {
	flag.StringVar(&DBFILE, "database_file", "bolt-kvb.db", "specify a filename for the database (BoltDB: https://github.com/boltdb/bolt)")
	flag.IntVar(&WPORT, "web_port", 8080, "specify the port to listen on for web connections")
	flag.IntVar(&BPORT, "backup_port", 8090, "specify the port to listen on for backups of the database file")
	flag.BoolVar(&BACKUPS, "backups", false, "specify whether the backup port should be enabled (defaults to false)")
}

func main() {
	finish := make(chan bool)

	flag.Parse()

	db, err := bolt.Open(DBFILE, 0600, nil)
	ce(err)
	DB = db //Global Pointer
	defer db.Close()

	initdb()

	webserver := http.NewServeMux()
	webserver.HandleFunc("/", rootHandler)
	webserver.HandleFunc("/b", rootHandler)
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
