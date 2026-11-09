package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	db "github.com/y0f/gophling/v2/internal/gen/db"
	"github.com/y0f/gophling/v2/internal/platform/problem"
)

// Claim ties a create to a client-supplied Idempotency-Key. The key is
// inserted in the same transaction as the created row, so a concurrent retry
// that loses the race is rolled back rather than creating a duplicate.
type Claim struct {
	Key  string
	Hash string
	TTL  time.Duration
}

// ErrReserved means another request already claimed the key; the caller
// should replay the stored response.
var ErrReserved = errors.New("idempotency key already reserved")

// maxKeyLen bounds the header, which is stored as the row key.
const maxKeyLen = 255

// subjectEscaper removes ':' from the subject, so the first ':' in a row key
// always ends the subject: subject "a" with key "b:c" and subject "a:b" with
// key "c" do not share the row key "a:b:c". Subjects without ':' or '%' keep
// the row key they had before the escaping was added.
var subjectEscaper = strings.NewReplacer("%", "%25", ":", "%3A")

// RowKey is the stored key for a caller's Idempotency-Key. The stored key is a
// global primary key, so it is scoped to the caller: two principals picking
// the same Idempotency-Key value must not be served each other's response.
func RowKey(subject, key string) string {
	return subjectEscaper.Replace(subject) + ":" + key
}

// NewClaim validates key and subject and builds the claim. On failure it has
// written a 400 and returns false. Every token without a "sub" claim would
// share one namespace, so such a token is refused rather than mixed up.
func (s *Store) NewClaim(w http.ResponseWriter, r *http.Request, subject, key string, body []byte) (Claim, bool) {
	if len(key) > maxKeyLen {
		problem.Status(w, r, http.StatusBadRequest, "Idempotency-Key exceeds 255 characters")
		return Claim{}, false
	}
	if subject == "" {
		problem.Status(w, r, http.StatusBadRequest, "Idempotency-Key requires a token with a subject (sub) claim")
		return Claim{}, false
	}
	return Claim{
		Key:  RowKey(subject, key),
		Hash: Hash([]byte(subject), []byte(r.URL.Path), body),
		TTL:  s.ttl,
	}, true
}

// Replay writes the stored response for the claim when one exists. Location is
// locationPrefix plus the stored body's JSON "id" field, and is omitted when
// the body has none. It returns true when it has written a response (a replay,
// a 409, or a 500 on lookup failure).
func (s *Store) Replay(w http.ResponseWriter, r *http.Request, claim Claim, locationPrefix string) bool {
	record, err := s.Lookup(r.Context(), claim.Key, claim.Hash)
	switch {
	case errors.Is(err, ErrConflict):
		problem.Status(w, r, http.StatusConflict, "idempotency key already used for a different request")
		return true
	case err != nil:
		slog.ErrorContext(r.Context(), "idempotency lookup", slog.Any("error", err))
		problem.Status(w, r, http.StatusInternalServerError, "internal server error")
		return true
	case record != nil:
		writeRecord(w, record, locationPrefix)
		return true
	default:
		return false
	}
}

func writeRecord(w http.ResponseWriter, record *Record, locationPrefix string) {
	var probe struct {
		ID string `json:"id"`
	}
	if json.Unmarshal(record.Body, &probe) == nil && probe.ID != "" {
		w.Header().Set("Location", locationPrefix+probe.ID)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Idempotency-Replayed", "true")
	w.WriteHeader(record.Status)
	// A previously stored JSON response, served back as application/json.
	_, _ = w.Write(record.Body) //#nosec G705
}

// Put stores the response for the claim inside tx. It returns ErrReserved when
// a concurrent request already holds the key, so tx rolls back.
func Put(ctx context.Context, tx pgx.Tx, claim Claim, status int, response any) error {
	body, err := json.Marshal(response)
	if err != nil {
		return err
	}
	rows, err := db.New(tx).PutIdempotencyKey(ctx, db.PutIdempotencyKeyParams{
		Key:            claim.Key,
		RequestHash:    claim.Hash,
		ResponseStatus: int32(status), //#nosec G115 -- an HTTP status fits in int32
		ResponseBody:   body,
		ExpiresAt:      time.Now().Add(claim.TTL),
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrReserved
	}
	return nil
}
