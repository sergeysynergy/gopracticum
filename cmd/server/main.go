package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}
