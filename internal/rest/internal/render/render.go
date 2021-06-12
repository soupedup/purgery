// Package render implements rendering helpers.
package render

import "net/http"

// NoContent writes a HTTP 204 No Content response to the given ResponseWriter.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// UnprocessableEntity writes a HTTP 422 Unprocessable Entity response to the
// given ResponseWriter.
func UnprocessableEntity(w http.ResponseWriter) {
	code(w, http.StatusUnprocessableEntity)
}

// InternalServerError writes a HTTP 500 Internal Server Error response to the
// given ResponseWriter.
func InternalServerError(w http.ResponseWriter) {
	code(w, http.StatusInternalServerError)
}

// ServiceUnavailable writes a HTTP 503 Service Unavailable response to the
// given ResponseWriter.
func ServiceUnavailable(w http.ResponseWriter) {
	code(w, http.StatusServiceUnavailable)
}

// Unauthorized writes a HTTP 401 Unauthorized response to the given
// ResponseWriter.
func Unauthorized(w http.ResponseWriter) {
	code(w, http.StatusUnauthorized)
}

func code(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}
