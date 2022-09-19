package handlers

import (
	"log"
	"net/http"
)

func (h *Handler) ping(w http.ResponseWriter, r *http.Request) {
	//if h.dbStorer == nil {
	//	h.errorJSON(w, r, "database not plugged in", http.StatusInternalServerError)
	//	return
	//}

	if err := h.uc.Ping(); err != nil {
		h.errorJSON(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Successfully connected to database")
	w.WriteHeader(http.StatusOK)
}
