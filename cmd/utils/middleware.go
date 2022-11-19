package utils

import (
	"net/http"
	"strings"
)

func NoFileListing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") || r.URL.Path == "" {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
