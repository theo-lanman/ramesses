package main

import (
	"log"
	"time"

	"github.com/boltdb/bolt"
	"github.com/theo-lanman/sidecar/context"
	"github.com/theo-lanman/sidecar/forwarder"
	"github.com/theo-lanman/sidecar/receiver"
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
		if _, err := tx.CreateBucketIfNotExists(bucketName); err != nil {
			return err
		}
		return nil
	})

	context := context.New(bucketName, db)

	go forwarder.Start(context)

	httpErrors := make(chan error)
	go receiver.Start(context, httpErrors)
	// Begin waiting on the first http error we hit
	// Onc we get one, log fatal, exit 1
	log.Fatal(<-httpErrors)
}
