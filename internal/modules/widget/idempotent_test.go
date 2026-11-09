package widget

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/y0f/gophling/v2/internal/auth"
	db "github.com/y0f/gophling/v2/internal/gen/db"
	"github.com/y0f/gophling/v2/internal/idempotency"
)

func TestReplayChecksPermissionBeforeAccessingStore(t *testing.T) {
	h := NewHandler(NewService(&fakeRepo{}))
	// No idempotency store: denied requests must not reach it.
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/v1/widgets", nil)
	h.createIdempotent(w, r, auth.Principal{Subject: "alice"}, Input{Name: "x"}, []byte(`{"name":"x"}`), "existing-key")
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
	if w.Header().Get("Idempotency-Replayed") != "" {
		t.Fatal("denied request was replayed")
	}
}

func (f *fakeRepo) CreateIdempotent(_ context.Context, in Input, _ idempotency.Claim) (db.Widget, error) {
	f.createCalled = true
	return db.Widget{ID: uuid.New(), Name: in.Name, Status: in.Status}, nil
}

func TestServiceCreateIdempotentRequiresPermission(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	_, err := svc.CreateIdempotent(context.Background(), auth.Principal{}, Input{Name: "x"}, idempotency.Claim{Key: "k"})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("got %v, want ErrForbidden", err)
	}
	if repo.createCalled {
		t.Error("repository CreateIdempotent should not run when forbidden")
	}
}
