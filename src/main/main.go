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

	http.HandleFunc("/books/add", func(w http.ResponseWriter, r *http.Request) {
		var book ClassifyBookResponse
		var err error

		if book, err = find(r.FormValue("id")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if err = db.Ping(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		
		// Prepare statement for inserting data
		stmtIns, err := db.Prepare("INSERT INTO books (pk, title, author, id, classification) VALUES (?, ?, ?, ?, ?)") // ? = placeholder
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

		_, err = stmtIns.Exec(nil, book.BookData.Title, book.BookData.Author, book.BookData.ID,
				book.Classification.MostPopular)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	})

	fmt.Println(http.ListenAndServe(":7000", nil))
}

type ClassifySearchResponse struct {
	Results []SearchResult `xml:"works>work"`
}

type ClassifyBookResponse struct {
	BookData struct {
		Title string `xml:"title,attr"`
		Author string `xml:"author,attr"`
		Year string `xml:"hyr,attr"`
		ID string `xml:"owi,attr"`
	} `xml:"work"`
	Classification struct {
		MostPopular string `xml:"sfa,attr"`
	} `xml:"recommendations>ddc>mostPopular"`
}

func search(query string) ([]SearchResult, error) {
	var c ClassifySearchResponse
	link := "http://classify.oclc.org/classify2/Classify?&summary=true&title="
	body, err := classifyAPI(link + url.QueryEscape(query))

	fmt.Println(link + url.QueryEscape(query))
	
	if err != nil {
		return []SearchResult{}, err
	}

	err = xml.Unmarshal(body, &c)
	return c.Results, err
}

func find(id string) (ClassifyBookResponse, error) {
	var c ClassifyBookResponse
	link := "http://classify.oclc.org/classify2/Classify?&summary=true&owi="+ url.QueryEscape(id)
	body, err := classifyAPI(link)

	if err != nil {
		return ClassifyBookResponse{}, err 
	}

	err = xml.Unmarshal(body, &c)
	return c, err
}

func classifyAPI(url string) ([]byte, error) {
	var resp *http.Response
	var err error

	if resp, err = http.Get(url); err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}