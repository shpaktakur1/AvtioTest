package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

type wrapper func(fn http.HandlerFunc) http.HandlerFunc

var ErrMalformedDuration = errors.New("Malformed duration")

func respondWithAppError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func decodeJSONBody(body io.ReadCloser) (payload interface{}, err error) {

	decoder := json.NewDecoder(body)
	err = decoder.Decode(&payload)
	return
}

func processTTL(in string) (ttl time.Duration, err error) {
	if in == "" {
		in = "0s"
	}

	ttl, err = time.ParseDuration(in)
	if err != nil {
		err = ErrMalformedDuration
	}
	return
}

func Wrap(fn http.HandlerFunc, wrappers []wrapper) http.HandlerFunc {
	result := fn
	for _, wrapper := range wrappers {
		result = wrapper(result)
	}
	return result
}
