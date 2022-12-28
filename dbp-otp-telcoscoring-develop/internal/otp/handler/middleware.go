package handler

import (
	"encoding/json"
	"github.com/Nerzal/gocloak/v8"
	"github.com/rs/zerolog/log"
	"gitlab.apps.bcc.kz/digital-banking-platform/microservices/technical/dbp-otp-telcoscoring/configs"
	"net/http"
	"strings"
)

func Protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cfg := configs.InitEnvDBPostgre()
		w.Header().Set("Content-Type", "application/json")
		authHeader := r.Header.Get("Authorization")

		if len(authHeader) < 1 {
			w.WriteHeader(401)
			json.NewEncoder(w).Encode(UnauthorizedError())
			return
		}
		accessToken := strings.Split(authHeader, " ")[1]
		client := gocloak.NewClient(cfg.KeycloakHost)
		rptResult, err := client.RetrospectToken(r.Context(), accessToken, cfg.KeycloakClientId, cfg.KeycloakClientSecret,
			cfg.KeycloakRealm)
		if err != nil {
			log.Debug().Err(err)
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(BadRequestError(err.Error()))
			return
		}

		isTokenValid := *rptResult.Active

		if !isTokenValid {
			w.WriteHeader(401)
			json.NewEncoder(w).Encode(UnauthorizedError())
			return
		}

		// Our middleware logic goes here...
		next.ServeHTTP(w, r)
	})
}

type HttpError struct {
	Code    int    `json:"code"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func UnauthorizedError() HttpError {
	return HttpError{
		401,
		"Unauthorized",
		"You are not authorized to access this resource",
	}

}

func BadRequestError(message string) *HttpError {
	return &HttpError{
		400,
		"Bad Request",
		message,
	}

}
