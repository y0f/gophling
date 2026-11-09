package httpjson

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The body must match what an idempotent replay stores, which has no newline.
func TestWriteHasNoTrailingNewline(t *testing.T) {
	rec := httptest.NewRecorder()
	Write(rec, http.StatusCreated, map[string]string{"id": "1"})
	if rec.Code != http.StatusCreated || rec.Body.String() != `{"id":"1"}` {
		t.Errorf("got %d %q", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content type = %q", ct)
	}
}

func TestBodyError(t *testing.T) {
	for _, tc := range []struct {
		err        error
		wantStatus int
		wantDetail string
	}{
		{&http.MaxBytesError{Limit: 1}, http.StatusRequestEntityTooLarge, "request body exceeds the size limit"},
		{errors.New("boom"), http.StatusBadRequest, "invalid JSON body"},
	} {
		rec := httptest.NewRecorder()
		BodyError(rec, httptest.NewRequest(http.MethodPost, "/", nil), tc.err, "invalid JSON body")
		if rec.Code != tc.wantStatus || !strings.Contains(rec.Body.String(), tc.wantDetail) {
			t.Errorf("%v: got %d %s", tc.err, rec.Code, rec.Body)
		}
	}
}
