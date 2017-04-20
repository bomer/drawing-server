package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	// _ "github.com/lib/pq"
	"log"
	"net/http"
	// "time"
)

var schemas = [...]string{`
CREATE TABLE person (
    firstname text,
    lastname text,
    email text
);`,

	`
	CREATE TABLE drawing (
		id int(11) unsigned NOT NULL AUTO_INCREMENT,
	    username varchar(255),
	    drawing varchar(32000),
	    upvotes int(11) unsigned NOT NULL DEFAULT 0,
	    
	    PRIMARY KEY (id)
	);
`}

var db *sqlx.DB

type Person struct {
	Firstname string
	Lastname  string
	Email     string
}

type Drawing struct {
	Id       int
	Username string
	Drawing  string
	Upvotes  int
	// Date_Created  time.Time
	// Date_Modified time.Time
}

func setupDb() {
	db.MustExec("drop table IF EXISTS person")
	db.MustExec("drop table IF EXISTS drawing")

	db.MustExec(schemas[0])
	db.MustExec(schemas[1])

	for i := 0; i < 10; i++ {
		tx := db.MustBegin()
		tx.MustExec("INSERT INTO person (firstname, lastname, email) VALUES (?, ?, ?)", "Jason", "Moiron", "jmoiron@jmoiron.net")
		tx.MustExec("INSERT INTO person (firstname, lastname, email) VALUES (?, ?, ?)", "John", "Doe", "johndoeDNE@gmail.net")
		// tx.MustExec("INSERT INTO place (country, city, telcode) VALUES (?, ?, ?)", "United States", "New York", "1")
		// tx.MustExec("INSERT INTO place (country, telcode) VALUES (?, ?)", "Hong Kong", "852")
		// tx.MustExec("INSERT INTO place (country, telcode) VALUES (?, ?)", "Singapore", "65")
		// // Named queries can use structs, so if you have an existing struct (i.e. person := &Person{}) that you have populated, you can pass it in as &person
		tx.NamedExec("INSERT INTO person (firstname, lastname, email) VALUES (:firstname, :lastname, :email)", &Person{"Jane", "Citizen", "jane.citzen@example.com"})
		tx.Commit()
	}

}

func main() {
	// this Pings the database trying to connect, panics on error
	// use sqlx.Open() for sql.Open() semantics
	var err error
	db, err = sqlx.Connect("mysql", "root:password@/drawing")
	if err != nil {
		log.Fatalln(err)
	}
	db.Ping()
	fmt.Printf("Connected I think\n")

	setupDb()

	// Query the database, storing results in a []Person (wrapped in []interface{})
	people := []Person{}
	db.Select(&people, "SELECT * FROM person ORDER BY firstname ASC")
	// jason, john := people[0], people[1]

	// fmt.Printf("%#v\n%#v", jason, john)

	http.HandleFunc("/", showAll)
	http.HandleFunc("/add", addDrawing)
	http.HandleFunc("/person", showPerson)
	http.ListenAndServe(":3000", nil)

}

func addDrawing(w http.ResponseWriter, r *http.Request) {

	username := r.FormValue("username")
	drawing := r.FormValue("drawing")
	if username != "" && drawing != "" {
		// insert := &Drawing{Username: username, Drawing: drawing}
		tx := db.MustBegin()
		tx.MustExec("INSERT INTO drawing (username, drawing) VALUES (?, ?)", username, drawing)
		tx.Commit()
		fmt.Printf("Added user %v \n", username)
	}

}
func showAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	// people := []Person{}
	// db.Select(&people, "SELECT * FROM person")
	// json.NewEncoder(w).Encode(people)

	drawings := []Drawing{}
	db.Select(&drawings, "SELECT * FROM drawing")

	fmt.Printf("%v done", drawings)
	json.NewEncoder(w).Encode(drawings)
}

func showPerson(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, http.StatusText(400), 400)
		return
	}
	people := []Person{}
	db.Select(&people, "SELECT * FROM person where firstname LIKE ? ORDER BY firstname ASC limit 1 ", name)
	json.NewEncoder(w).Encode(people)
}
