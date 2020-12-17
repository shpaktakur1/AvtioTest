package rest

import "net/http"

type Authorizer interface {
	Authorize(string, string) bool
}

type BasicAuthorizer struct {
	Username string
	Password string
}

func (ba *BasicAuthorizer) Authorize(username string, password string) bool {
	return ba.Username == username && ba.Password == password
}

func auth(a Authorizer) wrapper {
	return func(fn http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user, password, ok := r.BasicAuth()
			if !ok || !a.Authorize(user, password) {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			fn(w, r)
		}
	}
}
