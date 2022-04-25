package handlers

import (
	"log"
	"net/http"
)

func (h *Handler) ping(w http.ResponseWriter, _ *http.Request) {
	if h.dbStorer == nil {
		h.errorJSON(w, "database not plugged in", http.StatusInternalServerError)
		return
	}

	if err := h.dbStorer.Ping(); err != nil {
		h.errorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Successfully connected to database")
	w.WriteHeader(http.StatusOK)
}
