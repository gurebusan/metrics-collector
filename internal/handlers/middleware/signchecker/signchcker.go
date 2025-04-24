package signchecker

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
)

func New(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			recievedHash := r.Header.Get("HashSHA256")
			if recievedHash != "" {
				body, err := io.ReadAll(r.Body)
				fmt.Println(string(body))
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error":"Failed to read request body"}`))
					return
				}
				r.Body = io.NopCloser(bytes.NewBuffer(body))

				expectedHash := createHash(body, key)
				fmt.Println(expectedHash)
				fmt.Println(recievedHash)
				if recievedHash != expectedHash {
					w.WriteHeader(http.StatusBadRequest)
					w.Write(body)
					return
				}
			}

			rec := &responseRecorder{
				ResponseWriter: w,
				body:           new(bytes.Buffer),
			}
			next.ServeHTTP(rec, r)

			signature := createHash(rec.body.Bytes(), key)
			w.Header().Set("HashSHA256", signature)
			w.Write(rec.body.Bytes())
		}
		return http.HandlerFunc(fn)
	}
}

func createHash(data []byte, key string) string {
	fmt.Println(key)
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

type responseRecorder struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}
