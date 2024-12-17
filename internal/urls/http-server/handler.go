package urlshandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/mashalin/multiplexer/internal/urls/dto"
)

const MaxURLs = 20

var errTooManyURLs = errors.New("too many URLs in request")

type UrlsService interface {
	Fetch(ctx context.Context, urls []string) ([]dto.ResponseData, error)
	FetchOne(ctx context.Context, url string) (string, error)
}

type Handler struct {
	service UrlsService
}

func New(service UrlsService) *Handler {
	return &Handler{
		service,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.fetch(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (h *Handler) fetch(w http.ResponseWriter, r *http.Request) {
	var request dto.RequestBody
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(request.URLs) > MaxURLs {
		http.Error(w, errTooManyURLs.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	result, err := h.service.Fetch(ctx, request.URLs)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
