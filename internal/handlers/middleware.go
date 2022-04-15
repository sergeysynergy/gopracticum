package handlers

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

func gzipDecompressor(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// распаковываем тело запроса, если оно сжато gzip
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				r.Body = gz
			}
			defer gz.Close()
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func gzipCompressor(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		gzw := w

		// создаём объект Writer с жатием, если клиент поддерживает gzip
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				log.Println("[ERROR] Failed to create gzip writer - ", err)
			} else {
				w.Header().Set("Content-Encoding", "gzip")
				// заменяем Writer на новый, с поддержкой gzip-сжатия
				gzw = gzipWriter{ResponseWriter: w, Writer: gz}
			}
			defer gz.Close()
		}

		next.ServeHTTP(gzw, r)
	}

	return http.HandlerFunc(fn)
}
