package handler

import (
	"awesomeProject2/internal/email"
	"awesomeProject2/internal/integrations/aws/s3"
	"awesomeProject2/internal/integrations/aws/ses"
	"awesomeProject2/internal/integrations/db"
	"awesomeProject2/internal/model"
	"awesomeProject2/internal/transaction"
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/joho/godotenv"
	"net/http"
	"os"
)

//go:generate mockgen -source=handler.go -destination=handler_mock.go -package=handler

type PgConfig struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type Config struct {
	AwsSesConfig AwsSesConfig `json:"aws_ses_config"`
}

type AwsConfig struct {
	Region string `json:"region"`
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

type AwsSesConfig struct {
	AwsConfig
	From string `json:"from"`
}

type Handler struct {
	service service
}

type service interface {
	ProcessCsv(context.Context) (*model.Summary, float64, error)
}

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

func NewHandler(service service) *Handler {
	return &Handler{
		service: service,
	}
}

func session(configs map[string]string) *Handler {
	db := db.InitPostgres(configs["host"], configs["database"], configs["user"], configs["password"])
	repository := transaction.NewRepository(db)
	emailService := email.NewService(ses.NewService(configs))
	s3Service := s3.NewS3Service(configs)

	return NewHandler(transaction.NewService(emailService, s3Service, repository))
}

func LambdaEvent() (events.APIGatewayProxyResponse, error) {
	ctx := context.Background()

	credentials, err := buildConfig()
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, err
	}

	h := session(credentials)
	_, runningBalance, err := h.service.ProcessCsv(ctx)
	if err != nil {
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       err.Error(),
			}, err
		}
	}

	fmt.Println(runningBalance)

	return events.APIGatewayProxyResponse{
		Body:       "Successful",
		StatusCode: http.StatusOK,
	}, nil
}
