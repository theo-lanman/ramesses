package receiver

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/theo-lanman/sidecar/context"
	"github.com/theo-lanman/sidecar/message"
	"io/ioutil"
	"log"
	"net/http"
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

// Start
// Starts a receiver
func Start(c *context.Context, errorQueue chan<- error) {
	s := newContextServeMux(c)
	s.handleContextFunc("/jobs/all", jobsAllGet)
	s.handleContextFunc("/jobs", jobsPost)

	log.Printf("Listening...")

	// blocks while serving; always returns a non-nil error
	err := http.ListenAndServe(":5050", s)
	errorQueue <- err
}

// Accepts a POST request, and attempts to write the request to the database
func jobsPost(c *context.Context, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "500 Could not read body", 500)
		}

		var jobId uint64
		c.DB.Batch(func(tx *bolt.Tx) error {
			bucket := tx.Bucket(c.BucketName)
			id, err := bucket.NextSequence()
			if err != nil {
				return err
			}

			msgBytes, err := json.Marshal(message.NewMessage(id, body))
			if err != nil {
				return err
			}

			err = bucket.Put(message.Itob(id), msgBytes)
			if err != nil {
				return err
			}
			jobId = id
			return nil
		})
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
				json.Unmarshal(v, &msg)
				fmt.Fprintf(w, "key=%v value=%v\n", binary.BigEndian.Uint64(k), string(msg.Body))
			}
			return nil
		})
		fmt.Fprintf(w, "done\n")
	default:
		http.Error(w, "405 method not allowed", 405)
	}
}
