package receiver

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/theo-lanman/sidecar/context"
	"github.com/theo-lanman/sidecar/message"
)

// contextHandlerFunc
// An HTTP handler func which takes a context.Context pointer as an additional argument
type contextHandlerFunc func(*context.Context, http.ResponseWriter, *http.Request)

// contextHandler
// Wraps a contextHandlerFunc and a context.Context and implements the http.Handler interface
type contextHandler struct {
	*context.Context
	handlerFunc contextHandlerFunc
}

func (h contextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handlerFunc(h.Context, w, r)
}

// contextServeMux
// An http.ServeMux with a context.Context, to simplify handling http with contextHandlerFuncs
type contextServeMux struct {
	*context.Context
	*http.ServeMux
}

func newContextServeMux(c *context.Context) *contextServeMux {
	return &contextServeMux{c, http.NewServeMux()}
}

func (s *contextServeMux) handleContextFunc(pattern string, handlerFunc contextHandlerFunc) {
	s.Handle(pattern, contextHandler{s.Context, handlerFunc})
}

// Starts a receiver
func Start(c *context.Context, errorQueue chan<- error) {
	s := newContextServeMux(c)
	s.handleContextFunc("/jobs", jobsPost)

	log.Printf("Listening...")

	// blocks while serving; always returns a non-nil error
	errorQueue <- http.ListenAndServe(":5050", s)
}

// jobsPost
// Handles a POST request, and attempts to write the request to the database
func jobsPost(c *context.Context, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("500 Could not read body: %v", err), 500)
			return
		}

		var jobId uint64
		err = c.DB.Batch(func(tx *bolt.Tx) error {
			bucket := tx.Bucket(c.BucketName)
			id, err := bucket.NextSequence()
			if err != nil {
				return err
			}

			msg := message.NewMessage(id, body)
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				return err
			}

			if err = bucket.Put(msg.IdBytes(), msgBytes); err != nil {
				return err
			}

			jobId = id
			return nil
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("500 Error writing job: %v", err), 500)
			return
		}

		log.Printf("Stored item id=%v", jobId)
	default:
		http.Error(w, "405 method not allowed", 405)
	}
}

func jobsAllGet(c *context.Context, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		c.DB.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket(c.BucketName)
			cursor := bucket.Cursor()
			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				var msg message.Message
				if err := json.Unmarshal(v, &msg); err != nil {
					return err
				}
				fmt.Fprintf(w, "key=%v value=%v\n", binary.BigEndian.Uint64(k), string(msg.Body))
			}
			return nil
		})
		fmt.Fprintf(w, "done\n")
	default:
		http.Error(w, "405 method not allowed", 405)
	}
}
