package idempotency

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRowKeyIsUnambiguous(t *testing.T) {
	if RowKey("a", "b:c") == RowKey("a:b", "c") {
		t.Error("subjects containing ':' share row keys with other subjects")
	}
	if RowKey("a%3Ab", "c") == RowKey("a:b", "c") {
		t.Error("an escaped subject aliases a raw one")
	}
	// Rows stored before the escaping stay readable for ordinary subjects.
	if got := RowKey("alice", "k:1"); got != "alice:k:1" {
		t.Errorf("row key = %q, want alice:k:1", got)
	}
}

func TestNewClaim(t *testing.T) {
	store := NewStore(nil, time.Hour)
	for _, tc := range []struct {
		name, subject, key, wantDetail string
	}{
		{"overlong key", "alice", strings.Repeat("k", maxKeyLen+1), "Idempotency-Key exceeds 255 characters"},
		// Keys are namespaced by subject, so a token without one would share a
		// key space with every other subject-less caller.
		{"empty subject", "", "k", "Idempotency-Key requires a token with a subject (sub) claim"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			if _, ok := store.NewClaim(rec, httptest.NewRequest(http.MethodPost, "/v1/widgets", nil), tc.subject, tc.key, nil); ok {
				t.Fatal("claim accepted")
			}
			if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), tc.wantDetail) {
				t.Errorf("got %d %s, want 400 %q", rec.Code, rec.Body, tc.wantDetail)
			}
		})
	}

	body := []byte(`{"name":"x"}`)
	claim, ok := store.NewClaim(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/v1/widgets", nil), "a:b", "k", body)
	if !ok {
		t.Fatal("valid claim refused")
	}
	// The stored key and hash must not change, or existing rows stop replaying.
	want := Claim{Key: RowKey("a:b", "k"), Hash: Hash([]byte("a:b"), []byte("/v1/widgets"), body), TTL: time.Hour}
	if claim != want {
		t.Errorf("claim = %+v, want %+v", claim, want)
	}
}

// Replay builds Location from the stored body's JSON "id", so a resource whose
// response has no id is replayed without one.
func TestWriteRecordLocation(t *testing.T) {
	for _, tc := range []struct{ body, want string }{
		{`{"id":"42","name":"x"}`, "/v1/widgets/42"},
		{`{"name":"x"}`, ""},
		{`not json`, ""},
	} {
		rec := httptest.NewRecorder()
		writeRecord(rec, &Record{Status: http.StatusCreated, Body: []byte(tc.body)}, "/v1/widgets/")
		if got := rec.Header().Get("Location"); got != tc.want {
			t.Errorf("%s: Location = %q, want %q", tc.body, got, tc.want)
		}
		if rec.Code != http.StatusCreated || rec.Body.String() != tc.body || rec.Header().Get("Idempotency-Replayed") != "true" {
			t.Errorf("%s: replay = %d %q replayed=%q", tc.body, rec.Code, rec.Body, rec.Header().Get("Idempotency-Replayed"))
		}
	}
}
