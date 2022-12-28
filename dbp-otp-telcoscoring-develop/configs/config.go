package configs

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"os"
)

//здесь происходит работа с конфигурационным файлом

type Configs struct {
	HostDB               string
	PortDB               string
	DB                   string
	UsernameDB           string
	PasswordDB           string
	SendAzimutUrl        string
	VerifyAzimutUrl      string
	LogLevel             string
	OtpTable             string
	Migrations           string
	SchemaVersionTable   string
	KeycloakClientId     string
	KeycloakClientSecret string
	KeycloakRealm        string
	KeycloakHost         string
}

func InitEnvDBPostgre() Configs {
	err := godotenv.Load()
	if err != nil {
		log.Error().Err(err)
	}
	return Configs{
		HostDB:               os.Getenv("POSTGRE_HOST"),
		PortDB:               os.Getenv("POSTGRE_PORT"),
		DB:                   os.Getenv("POSTGRE_DB"),
		UsernameDB:           os.Getenv("POSTGRE_USER"),
		PasswordDB:           os.Getenv("POSTGRE_PASSWORD"),
		SendAzimutUrl:        os.Getenv("SEND_AZIMUT_URL"),
		VerifyAzimutUrl:      os.Getenv("VERIFY_AZIMUT_URL"),
		LogLevel:             os.Getenv("LOG_LEVEL"),
		OtpTable:             os.Getenv("OTP_TABLE"),
		Migrations:           os.Getenv("MIGRATIONS"),
		SchemaVersionTable:   os.Getenv("SCHEMA_VERSION_TABLE"),
		KeycloakClientId:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		KeycloakClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
		KeycloakRealm:        os.Getenv("KEYCLOAK_REALM"),
		KeycloakHost:         os.Getenv("KEYCLOAK_HOST"),
	}

}
