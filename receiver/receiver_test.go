package receiver

import (
	"testing"

	"github.com/boltdb/bolt"
	"github.com/theo-lanman/ramesses/context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

var jobsPostTestCases = []struct {
	method               string
	requestBody          io.Reader
	expectedStatus       int
	expectedResponseBody string
}{
	// Wrong method
	{"GET", nil, 405, "405 method not allowed\n"},
	{"DELETE", nil, 405, "405 method not allowed\n"},
	{"PUT", strings.NewReader("something"), 405, "405 method not allowed\n"},
	// Success
	{"POST", strings.NewReader("something"), 201, "Stored item id=1\n"},
}

// TestJobsPost tests basic HTTP responses for the test cases in
// jobsPostTestCases.
func TestJobsPost(t *testing.T) {
	c := testContext(t)

	for _, testCase := range jobsPostTestCases {
		req, err := http.NewRequest(testCase.method, "/jobs", testCase.requestBody)
		if err != nil {
			t.Fatal("Could not instantiate POST request")
		}

		w := httptest.NewRecorder()
		jobsPost(c, w, req)

		if w.Code != testCase.expectedStatus {
			t.Errorf("Method %v: expected status code %v, got %v", testCase.method, testCase.expectedStatus, w.Code)
		}

		if w.Body.String() != testCase.expectedResponseBody {
			t.Errorf("Method %v: expected body '%v', got '%v'", testCase.method, testCase.expectedResponseBody, w.Body.String())
		}
	}

	cleanup(c, t)
}

// testContext returns a context suitable for testing.
func testContext(t *testing.T) *context.Context {
	bucketName := []byte("jobs_test")
	db := testDB(t)
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	return context.New(bucketName, db)
}

// testDB returns a test database.
func testDB(t *testing.T) *bolt.DB {
	db, err := bolt.Open("jobs_test.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	return db
}

// cleanup removes the test database.
func cleanup(c *context.Context, t *testing.T) {
	dbPath := c.DB.Path()

	err := c.DB.Close()
	if err != nil {
		t.Logf("Could not close db: %v", err)
	}

	err = os.Remove(dbPath)
	if err != nil {
		t.Logf("Could not remove db (%v): %v", dbPath, err)
	}
}
