package main

import (
	"flag"
	"fmt"
	"net/http"
	"urlshort"

	"go.etcd.io/bbolt"
)

func main() {
	file := flag.String("i", "", "Input file, yaml and json supported (ex. urls.yaml)")
	flag.Parse()
	mux := defaultMux()

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	mapHandler := urlshort.MapHandler(pathsToUrls, mux)

	var (
		handler http.HandlerFunc
		db      *bbolt.DB
		err     error
	)

	if *file != "" {
		handler, err = urlshort.FileHandler(*file, mapHandler)
		if err != nil {
			panic(err)
		}
	} else {
		db, err = bbolt.Open("my.db", 0600, nil)
		if err != nil {
			panic(err)
		}
		defer db.Close()

		err = db.Update(func(tx *bbolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("Default"))
			for path, redirect := range pathsToUrls {
				err = b.Put([]byte(path), []byte(redirect))
			}
			return err
		})
		if err != nil {
			panic(err)
		}
		handler = urlshort.DBHandler(db, mapHandler)
	}

	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", handler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
