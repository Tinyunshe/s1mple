package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
)

var (
	adminUsername     = "admin"
	adminUserPassword = "1qazXSW@"
)

func AuthFailed(w http.ResponseWriter, message string) {
	w.Header().Set("WWW-Authenticate", `Basic realm="My REALM"`)
	w.WriteHeader(401)
	w.Write([]byte(message))
}

func BasicAuth(nextFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(adminUsername))
			expectedPasswordHash := sha256.Sum256([]byte(adminUserPassword))

			usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1
			passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1

			if usernameMatch && passwordMatch {
				nextFunc(w, r)
				return
			}
			AuthFailed(w, "Server need basic auth")
		}
	}
}
