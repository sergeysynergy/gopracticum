package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (h *Handler) errorJson(w http.ResponseWriter, message string, statusCode int) {
	type errorJson struct {
		Error      string
		StatusCode int
	}
	e := errorJson{
		Error:      message,
		StatusCode: statusCode,
	}

	b, err := json.Marshal(e)
	if err != nil {
		msg := fmt.Sprintf(`{"Error": "Failed to marshal error - %s", "StatusCode": 500`, err)
		w.Write([]byte(msg))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(b)
}

func (h *Handler) errorJsonUnsupportedMediaType(w http.ResponseWriter) {
	h.errorJson(w, "Wrong content type - application/json needed", http.StatusUnsupportedMediaType)
}

func (h *Handler) errorJsonReadBodyFailed(w http.ResponseWriter, err error) {
	h.errorJson(w, "Failed to read request body - "+err.Error(), http.StatusInternalServerError)
}

func (h *Handler) errorJsonUnmarshalFailed(w http.ResponseWriter, err error) {
	h.errorJson(w, "Unmarshal JSON failed - "+err.Error(), http.StatusNotAcceptable)
}

func (h *Handler) errorJsonMarshalFailed(w http.ResponseWriter, err error) {
	h.errorJson(w, "Marshal JSON failed - "+err.Error(), http.StatusInternalServerError)
}
