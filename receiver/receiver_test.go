package receiver

import (
	"testing"

	"github.com/theo-lanman/sidecar/context"
	"net/http"
	"net/http/httptest"
)

func TestJobsPostWrongMethod(t *testing.T) {
	req, err := http.NewRequest("GET", "/jobs", nil)
	if err != nil {
		t.Fatal("Could not instantiate GET request")
	}

	w := httptest.NewRecorder()
	c := context.New(nil, nil)

	jobsPost(c, w, req)

	if w.Code != 405 {
		t.Fail()
	}
}
