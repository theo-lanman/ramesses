package main

import (
	"github.com/boltdb/bolt"
	"github.com/theo-lanman/sidecar/context"
	"github.com/theo-lanman/sidecar/forwarder"
	"github.com/theo-lanman/sidecar/receiver"
	"log"
	"time"
)

// This is a prototype of a http-listening sidecar for publishing to a message queue such as Kafka.
func main() {
	// Instiate db and initialize buckets if necessary
	bucketName := []byte("messages")
	db, err := bolt.Open("sidecar.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}
		return nil
	})

	context := context.New(bucketName, db)

	go forwarder.Start(context)

	httpErrors := make(chan error)
	go receiver.StartHTTPServer(context, httpErrors)

	for {
		select {
		// If the HTTP server dies, we die
		case err := <-httpErrors:
			log.Fatal(err)
		}
	}
}
