package handler

import (
	"encoding/json"
	"net/http"

	"github.com/NemCaBong/executify/internal/service"
)

type Handler struct {
	svc service.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Hello(w http.ResponseWriter, r *http.Request) {
	// Call service method
	message := h.svc.GetHelloMessage()

	// Respond with JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": message,
	})
}
