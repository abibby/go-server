package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zwzn/go-server/rest"
)

func greet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World! %s", time.Now())
}

var resource = `
name: test
fields:
  - name: id
    type: string
    nullable: false
    primary: true
  - name: foo
    type: string
    nullable: false
  - name: bar
    type: int
    nullable: true
`

func main() {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `CREATE TABLE IF NOT EXISTS test (
	id STRING NOT NULL PRIMARY KEY,
	foo TEXT(255) NOT NULL DEFAULT '',
	bar INT
);
		  `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	res, err := rest.LoadResource(db, []byte(resource))
	if err != nil {
		log.Fatal(err)
	}

	res.Route(r.PathPrefix("/test").Subrouter())
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
