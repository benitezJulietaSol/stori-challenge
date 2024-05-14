package handler

import (
	"github.com/joho/godotenv"
	"os"
	"stori-challenge/internal/email"
	"stori-challenge/internal/integrations/aws/s3"
	"stori-challenge/internal/integrations/aws/ses"
	"stori-challenge/internal/integrations/db"
	"stori-challenge/internal/transaction"
)

func buildConfig() (map[string]string, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	config, err := godotenv.Read(path + "/.env")
	if err != nil {
		return nil, err
	}
	return config, nil
}

func session(configs map[string]string) *Handler {
	db := db.InitPostgres(configs["host"], configs["database"], configs["user"], configs["password"])
	repository := transaction.NewRepository(db)
	emailService := email.NewService(ses.NewService(configs))
	s3Service := s3.NewS3Service(configs)

	return NewHandler(transaction.NewService(emailService, s3Service, repository))
}

func config() *Handler {
	credentials, err := buildConfig()
	if err != nil {
		if err != nil {
			panic(err)
		}
	}

	h := session(credentials)
	return h
}
