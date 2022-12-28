package services

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"gitlab.apps.bcc.kz/digital-banking-platform/microservices/technical/dbp-otp-telcoscoring/internal/otp/entity"
	"gitlab.apps.bcc.kz/digital-banking-platform/microservices/technical/dbp-otp-telcoscoring/internal/otp/repository"
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) SendOtp(otpItems entity.OtpSendRequest) (string, error) {
	log.Debug().Msgf("start SentOTP")
	var code string
	var err error
	_, status, count, err := s.repo.GetStatusAndCount(otpItems.PhoneNumber)
	if err != nil {
		if err.Error() == "no rows in result set" {
			err := s.repo.New(otpItems)
			if err != nil {
				return code, err
			}
			code, err = s.SendSMS(otpItems, count)
			if err != nil {
				return code, err
			}
			return code, nil
		}
		log.Error().Err(err)
		fmt.Println(err)
		return code, errors.New("Не удалось отправить смс-код, попробуйте позже")
	}
	switch status {
	case entity.StatusNotSent:
		log.Debug().Msgf("activityId: %d, phone: %s, status: %s, count: %d", otpItems.ActivityId, otpItems.PhoneNumber, status, count)
		code, err = s.SendSMS(otpItems, count)
		if err != nil {
			return code, err
		}
		return code, nil

	case entity.StatusVerified:
		log.Debug().Msgf("activityId: %d, phone: %s, status: %s, count: %d", otpItems.ActivityId, otpItems.PhoneNumber, status, count)
		return code, errors.New("По данному номеру телефона уже был отправлен и подтверждён смс-код")

	case entity.StatusNotVerified:
		log.Debug().Msgf("activityId: %d, phone: %s, status: %s, count: %d", otpItems.ActivityId, otpItems.PhoneNumber, status, count)
		if count > 3 {
			err := s.repo.SetStatus(otpItems.PhoneNumber, entity.StatusFailed, count+1)
			if err != nil {
				log.Error().Err(err)
				return code, err
			}
			return code, errors.New("Вы ввели неправильный смс-код более 3-х раз, пожалуйста попробуйте через 24 часа")
		}
		if count <= 3 {
			code, err = s.SendSMS(otpItems, count)
			if err != nil {
				return code, err
			}
		}
		return code, nil

	case entity.StatusFailed:
		log.Debug().Msgf("activityId: %d, phone: %s, status: %s, count: %d", otpItems.ActivityId, otpItems.PhoneNumber, status, count)
		//err := s.SendSMS(otpItems, count)
		//if err != nil {
		//	return err
		//}
		return code, errors.New("Пожалуйста попробуйте через 24 часа")
		///
	case entity.StatusSent:
		log.Debug().Msgf("activityId: %d, phone: %s, status: %s, count: %d", otpItems.ActivityId, otpItems.PhoneNumber, status, count)
		code, err = s.SendSMS(otpItems, count)
		if err != nil {
			return code, err
		}
		return code, nil
	case entity.StatusNew:
		log.Debug().Msgf("activityId: %d, phone: %s, status: %s, count: %d", otpItems.ActivityId, otpItems.PhoneNumber, status, count)
		if err != nil {
			code, err = s.SendSMS(otpItems, count)
			if err != nil {
				return code, err
			}
			return code, nil
		}
	}

	err = s.repo.New(otpItems)
	if err != nil {
		return code, err
	}
	code, err = s.repo.SendAzimut(otpItems.PhoneNumber)
	if err != nil {
		//todo write status NOT Sent to DB
		return code, err
	}
	log.Debug().Msgf("end SentOTP")
	return code, nil
}

func (s *Service) VerifyOtp(otpVerifyItems entity.OtpVerifyRequest) error {
	log.Debug().Msgf("start VerifyOtp, activityId: %d, Phone: %s, code: %s", otpVerifyItems.ActivityId, otpVerifyItems.PhoneNumber, otpVerifyItems.Code)
	defer log.Debug().Msgf("end VerifyOtp, activityId: %d, Phone: %s, code: %s", otpVerifyItems.ActivityId, otpVerifyItems.PhoneNumber, otpVerifyItems.Code)
	_, status, count, err := s.repo.GetStatusAndCount(otpVerifyItems.PhoneNumber)
	if err != nil {
		return errors.New("Не удалось отправить смс-код, попробуйте позже")
	}
	log.Debug().Msgf("status: %s, count: %d", status, count)
	switch status {
	case entity.StatusSent:
		log.Debug().Msgf("activityId: %d, phone: %s, status: %s, count: %d", otpVerifyItems.ActivityId, otpVerifyItems.PhoneNumber, status, count)
		err = s.VerifySMS(otpVerifyItems, count)
		if err != nil {
			return err
		}
		return nil

	case entity.StatusVerified:
		log.Debug().Msgf("activityId: %d, phone: %s, status: %s, count: %d", otpVerifyItems.ActivityId, otpVerifyItems.PhoneNumber, status, count)
		return errors.New("смс-код по данному номеру телефона уже был подтвержден")

	case entity.StatusFailed:
		log.Debug().Msgf("activityId: %d, phone: %s, status: %s, count: %d", otpVerifyItems.ActivityId, otpVerifyItems.PhoneNumber, status, count)
		return errors.New("Пройзошла ошибка при отправке или подтверждении смс-кода")
	case entity.StatusNotVerified:
		log.Debug().Msgf("activityId: %d, phone: %s, status: %s, count: %d", otpVerifyItems.ActivityId, otpVerifyItems.PhoneNumber, status, count)
		err = s.VerifySMS(otpVerifyItems, count)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (s Service) SendSMS(items entity.OtpSendRequest, count int) (string, error) {
	code, err := s.repo.SendAzimut(items.PhoneNumber)
	if err != nil {
		fmt.Println(err)
		count++
		err := s.repo.SetStatus(items.PhoneNumber, entity.StatusNotSent, count)
		if err != nil {
			fmt.Println(err, "---")
			log.Error().Err(err)
		}
		log.Error().Err(err)
		fmt.Println(err)
		return "", errors.New("Не удалось отправить смс-код, введите правильный номер телефона")
	}
	err = s.repo.SetStatus(items.PhoneNumber, entity.StatusSent, count)
	if err != nil {
		log.Error().Err(err)
	}
	return code, nil
}

func (s Service) VerifySMS(items entity.OtpVerifyRequest, count int) error {
	log.Debug().Msgf("start VerifySMS, activityId: %d, phone: %s, count: %d", items.ActivityId, items.PhoneNumber, count)
	defer log.Debug().Msgf("end VerifySMS, activityId: %d, phone: %s, count: %d", items.ActivityId, items.PhoneNumber, count)
	err := s.repo.VerifyAzimut(items.PhoneNumber, items.Code)
	if err != nil {
		log.Debug().Err(err)
		if count >= 3 {
			err := s.repo.SetStatus(items.PhoneNumber, entity.StatusFailed, count+1)
			if err != nil {
				log.Error().Err(err)
			}
			return errors.New("Вы превысили допустимый лимит ввода смс-кода, пожалуйста попробуйте через 24 часа")
		}
		err := s.repo.SetStatus(items.PhoneNumber, entity.StatusNotVerified, count+1)
		if err != nil {
			log.Error().Err(err)
		}
		return errors.New("Не удалось верифицировать смс-код, проверьте смс-код и номер телефона")
	}
	err = s.repo.SetStatus(items.PhoneNumber, entity.StatusVerified, count)
	if err != nil {
		log.Error().Err(err)
	}
	return nil
}
