package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	myhttp "github.com/aleks0ps/gophermart/internal/app/http"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

func Gzipper() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fnDec := func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			var bReader io.ReadCloser = gz
			defer gz.Close()
			dR, err := http.NewRequestWithContext(r.Context(), r.Method, r.URL.String(), bReader)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// Set the same content type
			dR.Header.Set("Content-Type", r.Header.Get("Content-Type"))
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				typeCode := myhttp.GetContentTypeCode(r.Header.Get("Content-Type"))
				switch typeCode {
				case myhttp.CTypePlain, myhttp.CTypeHTML:
					eW, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					// Pass decoded request
					// Return encoded respose
					next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: eW}, dR)
				default:
					next.ServeHTTP(w, dR)
				}
			} else {
				next.ServeHTTP(w, dR)
			}
		}
		return http.HandlerFunc(fnDec)
	}
}
