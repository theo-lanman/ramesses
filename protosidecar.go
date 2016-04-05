package main

import (
	"encoding/binary"
	"fmt"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const bucketName = "jobs"

// This is a prototype of a http-listening sidecar for publishing to a message queue such as Kafka.
func main() {
	// Instiate db and initialize buckets if necessary
	db, err := bolt.Open("sidecar.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		return nil
	})

	httpErrors := make(chan error)
	go StartHTTPServer(httpErrors, db)

	for {
		select {
		// If the HTTP server dies, we die
		case err := <-httpErrors:
			log.Fatal(err)
		}
	}
}

func StartHTTPServer(errorQueue chan<- error, db *bolt.DB) {
	// Accepts a POST request, and attempts to write the request to the database
	http.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "500 Could not read body", 500)
			}

			var jobId int
			db.Batch(func(tx *bolt.Tx) error {
				bucket := tx.Bucket([]byte(bucketName))
				id, err := bucket.NextSequence()
				if err != nil {
					return err
				}

				err = bucket.Put(itob(id), body)
				if err != nil {
					return err
				}
				jobId = int(id)
				return nil
			})
			log.Printf("Stored item id=%v", jobId)
		case "GET":
			db.View(func(tx *bolt.Tx) error {
				bucket := tx.Bucket([]byte(bucketName))
				cursor := bucket.Cursor()
				for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
					fmt.Fprintf(w, "key=%v value=%v\n", binary.BigEndian.Uint64(k), string(v))
				}
				return nil
			})
			fmt.Fprintf(w, "done\n")
		default:
			http.Error(w, "405 method not allowed", 405)
		}
	})

	// TODO: configurability
	log.Printf("Listening...")
	err := http.ListenAndServe(":5050", nil)
	if err != nil {
		errorQueue <- err
	}
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
