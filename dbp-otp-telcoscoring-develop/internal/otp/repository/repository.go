package repository

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
	"gitlab.apps.bcc.kz/digital-banking-platform/microservices/technical/dbp-otp-telcoscoring/configs"
	"gitlab.apps.bcc.kz/digital-banking-platform/microservices/technical/dbp-otp-telcoscoring/internal/otp/entity"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Repository struct {
	pool *pgxpool.Pool
	cfg  configs.Configs
}

func NewRepository(pool *pgxpool.Pool, cfg configs.Configs) *Repository {
	return &Repository{
		pool: pool,
		cfg:  cfg,
	}
}

func (s *Repository) New(record entity.OtpSendRequest) error {
	psqls := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	t := time.Now().Unix()
	f := psqls.Insert("otpbcc").Columns("fail_count", "otp_status", "activity_id", "phone", "created_at", "updated_at")
	f = f.Values(0, entity.StatusNew, record.ActivityId, record.PhoneNumber, t, t)
	query, arqs, err := f.ToSql()
	fmt.Println(query)
	_, err = s.pool.Exec(context.Background(), query, arqs...)
	if err != nil {
		log.Error().Err(err).Msg("error Exec")
		return err
	}
	return nil
}

func (s Repository) SendAzimut(phone string) (string, error) {
	log.Debug().Msgf("Start SendAzimut")
	payloadB, err := json.Marshal(map[string]interface{}{
		"phone":           phone,
		"messageTemplate": "Ваш код: {0}",
		"otpLength":       6,
		"otpDurationSec":  300,
	})
	if err != nil {
		log.Error().Err(err).Msg("error Marshal")
		return "", err
	}
	payload := bytes.NewReader(payloadB)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}
	request, err := http.NewRequest(http.MethodPost, s.cfg.SendAzimutUrl, payload)
	if err != nil {
		log.Error().Err(err).Msg("error NewRequesr")
		return "", err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Error().Err(err).Msg("error Client DO")
		return "", err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error().Err(err).Msgf("error readAll")
		return "", err
	}
	switch response.StatusCode {
	case 200:
		var o entity.OtpSendResponseAzimut
		err := json.Unmarshal(body, &o)
		if err != nil {
			log.Error().Err(err).Msg("error Unmarshal")
			return "", err
		}
		fmt.Println(string(body))
		log.Debug().Msgf("end SendAzimut")
		return o.Code, err
	default:
		log.Debug().Msgf("end SendAzimut")
		return "", errors.New("During verification stage, we got error from Azimut")
	}
}

func (s Repository) VerifyAzimut(phone string, code string) error {
	log.Debug().Msgf("Start VerifyAzimut")
	payloadD, err := json.Marshal(map[string]interface{}{
		"phone": phone,
		"code":  code,
	})
	if err != nil {
		log.Error().Err(err).Msg("error Marshal")
		return err
	}
	payload := bytes.NewReader(payloadD)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}
	log.Debug().Msgf("URL from ENV: %s", os.Getenv("VERIFY_AZIMUT_URL"))
	log.Debug().Msg(s.cfg.VerifyAzimutUrl)
	log.Debug().Msg(string(payloadD))
	req, err := http.NewRequest(http.MethodPost, s.cfg.VerifyAzimutUrl, payload)
	if err != nil {
		log.Error().Err(err).Msg("error newRequest")
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("error ClientDo")
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error().Err(err).Msg("error ReadAll")
		return err
	}
	switch res.StatusCode {
	case 200:
		var o entity.OtpVerifyResponseAzimut
		err := json.Unmarshal(body, &o)
		if err != nil {
			log.Error().Err(err).Msg("error Unmarshall")
			return err
		}
		log.Debug().Msgf("End VerifyAzimut")
		return nil
	default:
		log.Debug().Msgf("End VerifyAzimut")
		return errors.New("Введенный номер неверный или код уже был активирован")
	}
}

func (r Repository) GetStatusAndCount(phone string) (int, string, int, error) {
	fmt.Println(phone)
	query := `select activity_id, otp_status, fail_count
		from otpbcc
		where phone =$1 and to_timestamp(updated_at) >= NOW() + INTERVAL '- 24 hours';`
	row := r.pool.QueryRow(context.Background(), query, phone)
	var count int
	var status string
	var activityId int
	err := row.Scan(&activityId, &status, &count)
	if err != nil {
		log.Error().Err(err).Msg("error Scan")
		return 0, "", 0, err
	}
	return activityId, status, count, nil
}

func (r Repository) SetStatus(phone string, status string, count int) error {

	switch status {
	case entity.StatusVerified:
		query := `update otpbcc set otp_status = $1 where phone = $2;`
		_, err := r.pool.Exec(context.Background(), query, status, phone)
		if err != nil {
			return err
		}
	case entity.StatusSent:
		query := `update otpbcc set otp_status = $1 where phone = $2;`
		_, err := r.pool.Exec(context.Background(), query, status, phone)
		if err != nil {
			return err
		}
	case entity.StatusNotVerified:
		query := `update otpbcc set otp_status = $1, fail_count = $2 where phone = $3;`
		_, err := r.pool.Exec(context.Background(), query, status, count, phone)
		if err != nil {
			return err
		}
	case entity.StatusNotSent:
		query := `update otpbcc set otp_status = $1, fail_count = $2 where phone = $3;`
		_, err := r.pool.Exec(context.Background(), query, status, count, phone)
		if err != nil {
			return err
		}
	case entity.StatusFailed:
		query := `update otpbcc set otp_status = $1, fail_count = $2 where phone = $3;`
		_, err := r.pool.Exec(context.Background(), query, status, count, phone)
		if err != nil {
			return err
		}

	}
	return nil
}
