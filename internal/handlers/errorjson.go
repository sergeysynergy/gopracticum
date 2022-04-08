package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (h *Handler) errorJSON(w http.ResponseWriter, message string, statusCode int) {
	type errorJSON struct {
		Error      string
		StatusCode int
	}
	e := errorJSON{
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

func (h *Handler) errorJSONUnsupportedMediaType(w http.ResponseWriter) {
	h.errorJSON(w, "Wrong content type - application/json needed", http.StatusUnsupportedMediaType)
}

func (h *Handler) errorJSONReadBodyFailed(w http.ResponseWriter, err error) {
	h.errorJSON(w, "Failed to read request body - "+err.Error(), http.StatusInternalServerError)
}

func (h *Handler) errorJSONUnmarshalFailed(w http.ResponseWriter, err error) {
	h.errorJSON(w, "Unmarshal JSON failed - "+err.Error(), http.StatusNotAcceptable)
}

func (h *Handler) errorJSONMarshalFailed(w http.ResponseWriter, err error) {
	h.errorJSON(w, "Marshal JSON failed - "+err.Error(), http.StatusInternalServerError)
}
