package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pgray/kvb/db"
	"github.com/russross/blackfriday"
)

//Globals
var DB *bolt.DB

//Flags
type Config struct {
	DatabaseFile string
	WebPort      int
	BackupPort   int
	Backups      bool
}

var config Config

func ce(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request, section string, title string) {
	fmt.Println(time.Now(), "saveHandler: ", section, title)
	body := r.FormValue("body")
	db.SavePage(DB, section, db.Page{Title: title, Body: []byte(body)})
	fmt.Println("save: ", section, title)
	http.Redirect(w, r, "/b/"+section+"/"+title, http.StatusFound)
}

func editHandler(w http.ResponseWriter, r *http.Request, section string, title string) {
	fmt.Println("editHandler: ", section, title)
	if title == "" {
		title = "root"
	}

	p := db.LoadPage(DB, section, title)
	t, err := template.New("edit").Parse(editHTML)
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
		t, err := template.New("browse").Parse(sectionHTML)
		ce(err)
		posts := db.Posts(DB, section)
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

	p := db.LoadPage(DB, section, title)

	if p.Body == nil {
		p.Body = append(p.Body, []byte("Sorry, that page does not exist")...)
	}

	t, err := template.New("browse").Parse(browseHTML)
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
	t, err := template.New("root").Parse(rootHTML)
	ce(err)
	test := db.Sections(DB)
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
	flag.StringVar(&config.DatabaseFile, "database_file", "bolt-kvb.db", "specify a filename for the database")
	flag.IntVar(&config.WebPort, "web_port", 8080, "specify the port to listen on for web connections")
	flag.IntVar(&config.BackupPort, "backup_port", 8090, "specify the port to listen on for backups of the database")
	flag.BoolVar(&config.Backups, "backups", false, "specify whether the backup port should be enabled")
	flag.Parse()
}

func main() {
	finish := make(chan bool)
	db, err := bolt.Open(config.DatabaseFile, 0600, nil)
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

	fmt.Println("Using database file: ", config.DatabaseFile)

	go func() {
		fmt.Println("Web server starting on port: ", config.WebPort)

		err := http.ListenAndServe(":"+strconv.Itoa(config.WebPort), webserver)

		if err != nil {
			log.Fatal("webserver ListenAndServe: ", err)
		}
	}()

	if config.Backups == true {
		go func() {
			fmt.Println("Backups available at /backup on port: ", config.BackupPort)

			err := http.ListenAndServe(":"+strconv.Itoa(config.BackupPort), backupserver)

			if err != nil {
				log.Fatal("backupserver ListenAndServe: ", err)
			}
		}()
	}

	<-finish
}
