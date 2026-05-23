package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/ninosh210804/lumiera/apps/server/internal/domain"
)

type envelope struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Data: data})
}

func Error(w http.ResponseWriter, err error) {
	var appErr *domain.AppError
	code := http.StatusInternalServerError
	msg := "internal server error"

	if e, ok := err.(*domain.AppError); ok {
		appErr = e
		code = appErr.Code
		msg = appErr.Message
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(envelope{Error: msg})
}
