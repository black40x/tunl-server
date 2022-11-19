package utils

import (
	"net/http"
	"strings"
)

func HasCookie(r *http.Request, name string) bool {
	for _, cookie := range r.Cookies() {
		if cookie.Name == name {
			return true
		}
	}
	return false
}

func IsBrowserRequest(r *http.Request) bool {
	return strings.Index(r.Header.Get("User-Agent"), "Mozilla") != -1
}
