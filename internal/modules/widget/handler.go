package widget

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/y0f/gophling/v2/internal/auth"
	api "github.com/y0f/gophling/v2/internal/gen/api"
	db "github.com/y0f/gophling/v2/internal/gen/db"
	"github.com/y0f/gophling/v2/internal/idempotency"
	"github.com/y0f/gophling/v2/internal/platform/httpjson"
	"github.com/y0f/gophling/v2/internal/platform/problem"
)

// Handler adapts HTTP to the widget service. It implements the generated
// api.ServerInterface for the widget operations.
type Handler struct {
	api.Unimplemented // Unrelated resource operations are implemented by the composition root.
	service           *Service
	//forge:begin idempotency
	idempotency *idempotency.Store
	//forge:end idempotency
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListWidgets(w http.ResponseWriter, r *http.Request, params api.ListWidgetsParams) {
	limit, offset := int32(20), int32(0)
	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}
	items, total, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := api.WidgetList{
		Items:      make([]api.Widget, 0, len(items)),
		Pagination: api.Pagination{Limit: limit, Offset: offset, Total: total},
	}
	for _, item := range items {
		out.Items = append(out.Items, toAPIWidget(item))
	}
	httpjson.Write(w, http.StatusOK, out)
}

func (h *Handler) CreateWidget(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpjson.BodyError(w, r, err, "cannot read request body")
		return
	}
	var input api.WidgetInput
	if err := json.Unmarshal(body, &input); err != nil {
		problem.Status(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}
	actor, _ := auth.PrincipalFrom(r.Context())
	in := inputFromAPI(input)

	//forge:begin idempotency
	if key := r.Header.Get("Idempotency-Key"); h.idempotency != nil && key != "" {
		h.createIdempotent(w, r, actor, in, body, key)
		return
	}
	//forge:end idempotency
	created, err := h.service.Create(r.Context(), actor, in)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeCreated(w, created)
}

func writeCreated(w http.ResponseWriter, created db.Widget) {
	w.Header().Set("Location", "/v1/widgets/"+created.ID.String())
	httpjson.Write(w, http.StatusCreated, toAPIWidget(created))
}

func (h *Handler) GetWidget(w http.ResponseWriter, r *http.Request, id api.WidgetID) {
	found, err := h.service.Get(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpjson.Write(w, http.StatusOK, toAPIWidget(found))
}

func (h *Handler) UpdateWidget(w http.ResponseWriter, r *http.Request, id api.WidgetID) {
	var input api.WidgetInput
	// The router's SingleJSONDocument middleware has already rejected bodies
	// with trailing data after the JSON document.
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpjson.BodyError(w, r, err, "invalid JSON body")
		return
	}
	actor, _ := auth.PrincipalFrom(r.Context())
	updated, err := h.service.Update(r.Context(), actor, id, inputFromAPI(input))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpjson.Write(w, http.StatusOK, toAPIWidget(updated))
}

func (h *Handler) DeleteWidget(w http.ResponseWriter, r *http.Request, id api.WidgetID) {
	actor, _ := auth.PrincipalFrom(r.Context())
	if err := h.service.Delete(r.Context(), actor, id); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		problem.Status(w, r, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrNotFound):
		problem.Status(w, r, http.StatusNotFound, "widget not found")
	case errors.Is(err, ErrForbidden):
		problem.Status(w, r, http.StatusForbidden, "you do not have permission to perform this action")
	default:
		slog.ErrorContext(r.Context(), "widget handler error", slog.Any("error", err))
		problem.Status(w, r, http.StatusInternalServerError, "internal server error")
	}
}

func toAPIWidget(w db.Widget) api.Widget {
	return api.Widget{
		Id:          w.ID,
		Name:        w.Name,
		Description: w.Description,
		Status:      api.WidgetStatus(w.Status),
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
}

func inputFromAPI(in api.WidgetInput) Input {
	description := ""
	if in.Description != nil {
		description = *in.Description
	}
	status := string(api.Active)
	if in.Status != nil {
		status = string(*in.Status)
	}
	return Input{Name: in.Name, Description: description, Status: status}
}
