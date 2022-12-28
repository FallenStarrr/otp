package handler

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"gitlab.apps.bcc.kz/digital-banking-platform/microservices/technical/dbp-otp-telcoscoring/internal/otp/entity"
	"gitlab.apps.bcc.kz/digital-banking-platform/microservices/technical/dbp-otp-telcoscoring/internal/otp/services"
	"io"
	"net/http"
)

type Handler struct {
	serv *services.Service
}

func NewHandler(serv *services.Service) *Handler {
	return &Handler{
		serv: serv,
	}
}

func (h *Handler) SentOTP(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("start SentOTP")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-api-version", "1.0")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("ReadAll error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var record entity.OtpSendRequest
	err = json.Unmarshal(body, &record)
	if err != nil {
		log.Error().Err(err).Msg("Unmarshal error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	code, err := h.serv.SendOtp(record)
	if err != nil {
		log.Error().Err(err).Msg("SendOtp request error")
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	if code == "" {
		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, map[string]interface{}{"success": true, "error": nil})
	}
	if code != "" {
		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, map[string]interface{}{"success": true, "code": code, "error": nil})
	}
	log.Debug().Msgf("end SentOTP")

}

func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("start Verify")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-api-version", "1.0")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("ReadAll error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var record entity.OtpVerifyRequest
	err = json.Unmarshal(body, &record)
	if err != nil {
		log.Error().Err(err).Msg("Unmarshal error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.serv.VerifyOtp(record)
	if err != nil {
		fmt.Println(err, "Handler, VerifyOtp request error")
		log.Error().Err(err).Msg("VerifyOtp request error")
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{"success": true, "error": nil})
	log.Debug().Msgf("end Verify")
}
