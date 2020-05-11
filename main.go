package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// application is our app struct with config
type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	db       *sql.DB
}

// openDB opens a database connection
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// main is main app function
func main() {
	// create loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// read flags
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")
	flag.Parse()

	// open connection to db
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		errorLog.Fatal(err)
	}
	infoLog.Println("Pinged database successfully!")

	// populate config
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		db:       db,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()

	if err != nil {
		errorLog.Fatal(err)
	}
}

func (app *application) UpdatePaint() {

	// delete from paint where province <> 'nb' or products not like '%paint%';

	dat, err := ioutil.ReadFile("./formatted_data.json")
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	var in []PaintData

	err = json.Unmarshal(dat, &in)
	if err != nil {
		log.Fatal(err)
	}

	for _, x := range in {
		_, err := app.InsertPaintDepot(x)
		if err != nil {
			app.errorLog.Println(err)
		}
	}
	fmt.Println("Done")
}
