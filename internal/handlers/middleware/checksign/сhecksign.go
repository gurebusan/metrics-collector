package checksign

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

const hashHeader = "HashSHA256"

func New(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r) // 1!
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			r.Body = io.NopCloser(bytes.NewBuffer(body))

			receivedHash := r.Header.Get(hashHeader)
			if receivedHash == "" {
				http.Error(w, "HashSHA256 header is missing", http.StatusBadRequest)
				return
			}

			expectedHash := createHash(body, key)
			if receivedHash != expectedHash {
				http.Error(w, "Invalid HashSHA256", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn) // 2!
	}
}

func createHash(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
