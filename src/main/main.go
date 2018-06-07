package main

import (
	"io/ioutil"
	"fmt"
	"net/http"
	"html/template"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	
	"encoding/json"
	"net/url"
	"encoding/xml"
)

type Page struct {
	Name string
	DBStatus bool
}

type SearchResult struct {
	Title string `xml:"title,attr"`
	Author string `xml:"author,attr"`
	Year string `xml:"hyr,attr"`
	ID string `xml:"owi,attr"`
}

func main() {
	templates := template.Must(template.ParseFiles("templates/index.html"))

	// Create the database handle, confirm driver is present
	// db, err := sql.Open("mysql", "user:password@/dbname")
	db, err := sql.Open("mysql", "scott:tiger@/demo-dev")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	// Server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := Page{Name: "3Dcoder"}
		if name := r.FormValue("name"); name != "" {
			p.Name = name
		}
		p.DBStatus = db.Ping() == nil

		if err := templates.ExecuteTemplate(w, "index.html", p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		var results []SearchResult
		var err error	
		var input string = r.FormValue("search")

		if results, err = search(input); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(results); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		
	})

	fmt.Println(http.ListenAndServe(":7000", nil))
}

type ClassifySearchResponse struct {
	Results []SearchResult `xml:"works>work"`
}

func search(query string) ([]SearchResult, error) {
	var resp *http.Response
	var err error
	var link string = "http://classify.oclc.org/classify2/Classify?&summary=true&title="
	var input string = link + url.QueryEscape(query)

	//fmt.Println(input)

	if resp, err = http.Get(input); err != nil {
		return []SearchResult{}, err
	}

	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return []SearchResult{}, err
	}

	var c ClassifySearchResponse
	err = xml.Unmarshal(body, &c)
	return c.Results, err
}