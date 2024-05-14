package handler

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	log "github.com/sirupsen/logrus"
	"net/http"
	"stori-challenge/internal/model"
)

//go:generate mockgen -source=handler.go -destination=handler_mock.go -package=handler

type Handler struct {
	service service
}

type service interface {
	ProcessCsv(context.Context) (*model.Summary, error)
}

func NewHandler(service service) *Handler {
	return &Handler{
		service: service,
	}
}

func ProxyLambdaEvent() (events.APIGatewayProxyResponse, error) {
	return config().LambdaEvent()
}

func (h *Handler) LambdaEvent() (events.APIGatewayProxyResponse, error) {
	ctx := context.Background()

	summary, err := h.service.ProcessCsv(ctx)
	if err != nil {
		log.WithContext(ctx).
			WithFields(log.Fields{"event": "lambda_event"}).
			Error(err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, err
	}

	body, err := json.Marshal(summary)
	if err != nil {
		log.WithContext(ctx).
			WithFields(log.Fields{"event": "lambda_event"}).
			Error(err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, err
	}

	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: http.StatusOK,
	}, nil
}
