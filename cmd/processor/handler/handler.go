package handler

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"stori-challenge/internal/model"
)

//go:generate mockgen -source=handler.go -destination=handler_mock.go -package=handler

type Handler struct {
	service service
}

type service interface {
	ProcessCsv(context.Context) (*model.Summary, float64, error)
}

func NewHandler(service service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) LambdaEvent() (events.APIGatewayProxyResponse, error) {
	ctx := context.Background()
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
