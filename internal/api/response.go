package api

import (
	"encoding/json"
	"net/http"
)

// Standard API response structures

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, ErrorResponse{
		Error: err.Error(),
	})
}

// WriteSuccess writes a success response
func WriteSuccess(w http.ResponseWriter, message string, data interface{}) {
	WriteJSON(w, http.StatusOK, SuccessResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// WriteCreated writes a 201 Created response
func WriteCreated(w http.ResponseWriter, message string, data interface{}) {
	WriteJSON(w, http.StatusCreated, SuccessResponse{
		Status:  "created",
		Message: message,
		Data:    data,
	})
}

// WriteNoContent writes a 204 No Content response
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
