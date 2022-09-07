package handlers

import (
	"compress/gzip"
	"crypto/rsa"
	"github.com/sergeysynergy/metricser/pkg/crypter"
	"io"
	"io/ioutil"
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

//func (c *Compressor) Handler(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		encoder, encoding, cleanup := c.selectEncoder(r.Header, w)
//
//		cw := &compressResponseWriter{
//			ResponseWriter:   w,
//			w:                w,
//			contentTypes:     c.allowedTypes,
//			contentWildcards: c.allowedWildcards,
//			encoding:         encoding,
//			compressable:     false, // determined in post-handler
//		}
//		if encoder != nil {
//			cw.w = encoder
//		}
//		// Re-add the encoder to the pool if applicable.
//		defer cleanup()
//		defer cw.Close()
//
//		next.ServeHTTP(cw, r)
//	})
//}

//func Compress(level int, types ...string) func(next http.Handler) http.Handler {
//	compressor := NewCompressor(level, types...)
//	return compressor.Handler
//}

//func Compress(level int) func(next http.Handler) http.Handler {
func decrypt(privateKey *rsa.PrivateKey) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//if r.Header.Get("Content-Encoding") == "gzip" {
			//	gz, err := gzip.NewReader(r.Body)
			//	if err != nil {
			//		http.Error(w, err.Error(), http.StatusInternalServerError)
			//	} else {
			//		r.Body = gz
			//	}
			//	defer gz.Close()
			//}

			if strings.Contains(r.Header.Get("Content-Encoding"), "crypted") && privateKey != nil {
				reqBody, err := ioutil.ReadAll(r.Body)
				if err != nil {
					log.Println("[ERROR] Failed to read body - ", err)
					return
				}
				//defer r.Body.Close()

				if privateKey != nil {
					reqBody, err = crypter.Decrypt(privateKey, reqBody)
					if err != nil {
						log.Println("[ERROR] Failed to decrypt body - ", err)
						return
					}
					log.Println("[INFO] Body has been successfully decrypted")
				}
				//r.Body.Read(reqBody)

				//reqBody, err := ioutil.ReadAll(r.Body)
				//if err != nil {
				//	log.Println("[ERROR] Failed to read body - ", err)
				//	return
				//}
				//defer r.Body.Close()

				//plainBody, err := crypter.Decrypt(privateKey, reqBody)
				//if err != nil {
				//	log.Println("[ERROR] Failed to decrypt body - ", err)
				//	return
				//}
				//r.Body.Read(plainBody)
				//log.Println("[INFO] Body has been successfully decrypted")

				//mcs := make([]metrics.Metrics, 0)
				//err = json.Unmarshal(reqBody, &mcs)
				//if err != nil {
				//	fmt.Println("EEEEEEEEERRRRRRRr", err)
				//	return
				//}
				//log.Printf("[DEBUG] %#v\n", mcs)
			}

			next.ServeHTTP(w, r)
		})
	}
}
