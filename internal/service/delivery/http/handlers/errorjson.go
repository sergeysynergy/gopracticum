package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func (h *Handler) errorJSON(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	reqID := middleware.GetReqID(r.Context())

	type errorJSON struct {
		Error      string
		StatusCode int
	}
	e := errorJSON{
		Error:      message,
		StatusCode: statusCode,
	}

	prefix := "[ERROR]"
	if reqID != "" {
		prefix = fmt.Sprintf("[%s] [ERROR]", reqID)
	}

	b, err := json.Marshal(e)
	if err != nil {
		msg := fmt.Sprintf(`{"Error": "Failed to marshal error - %s", "StatusCode": 500`, err)
		w.Write([]byte(msg))
		log.Println(prefix, msg)
		return
	}

	w.Header().Set("Content-Type", applicationJSON)
	w.WriteHeader(statusCode)
	w.Write(b)
	log.Println(prefix, e)
}

func (h *Handler) errorJSONUnsupportedMediaType(w http.ResponseWriter, r *http.Request) {
	h.errorJSON(w, r, "Wrong content type - application/json needed", http.StatusUnsupportedMediaType)
}

func (h *Handler) errorJSONReadBodyFailed(w http.ResponseWriter, r *http.Request, err error) {
	h.errorJSON(w, r, "Failed to read request body - "+err.Error(), http.StatusInternalServerError)
}

func (h *Handler) errorJSONUnmarshalFailed(w http.ResponseWriter, r *http.Request, err error) {
	h.errorJSON(w, r, "Unmarshal JSON failed - "+err.Error(), http.StatusNotAcceptable)
}

func (h *Handler) errorJSONMarshalFailed(w http.ResponseWriter, r *http.Request, err error) {
	h.errorJSON(w, r, "Marshal JSON failed - "+err.Error(), http.StatusInternalServerError)
}
