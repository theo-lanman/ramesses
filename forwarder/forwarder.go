package forwarder

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/theo-lanman/sidecar/context"
	"github.com/theo-lanman/sidecar/message"
)

func Start(c *context.Context) {
	workQueue := make(chan message.Batch)
	go StartWorker(c, workQueue)
	for {
		batch := message.NewBatch()
		items := 0
		c.DB.Update(func(tx *bolt.Tx) error {
			now := time.Now()
			visibilityHorizon := now.Add(-1 * time.Minute)
			bucket := tx.Bucket(c.BucketName)
			cursor := bucket.Cursor()
			// stop at 20 messages; probably this should be time-based with a large max size?
			for k, v := cursor.First(); k != nil && items < 20; k, v = cursor.Next() {
				//fmt.Printf("Read an item\n")
				var msg message.Message
				json.Unmarshal(v, &msg)
				// If this message hasn't been attempted recently, update it and put it in the batch.
				if msg.AttemptedAt.Before(visibilityHorizon) {
					msg.AttemptedAt = now
					batch = append(batch, msg)
					msgBytes, err := json.Marshal(msg)
					if err != nil {
						return err
					}
					if err = bucket.Put(k, msgBytes); err != nil {
						return err
					}
					items++
				}
			}
			return nil
		})
		fmt.Printf("Batch has %v items\n", len(batch))
		if len(batch) > 0 {
			workQueue <- batch
		}
		// TODO: real scheduling
		// If we didn't get a full batch, wait a little bit
		if len(batch) < 20 {
			time.Sleep(250 * time.Millisecond)
		}
	}
}

// TODO: move into its own package?
func StartWorker(c *context.Context, workQueue <-chan message.Batch) {
	for {
		select {
		case batch := <-workQueue:
			// TODO: handle failures--including partial success
			// send messages
			fmt.Printf("Starting batch\n")
			for _, msg := range batch {
				fmt.Printf("Processing message id=%v body=%v\n", msg.Id, string(msg.Body))
			}

			// delete messages from db
			c.DB.Update(func(tx *bolt.Tx) error {
				for _, msg := range batch {
					bucket := tx.Bucket(c.BucketName)
					if err := bucket.Delete(msg.IdBytes()); err != nil {
						return err
					}
				}
				return nil
			})
		}
	}
}
