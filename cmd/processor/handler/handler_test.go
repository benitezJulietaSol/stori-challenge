package handler

import (
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"stori-challenge/internal/model"
	"testing"
)

func TestLambdaEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	csvService := NewMockservice(ctrl)
	handler := NewHandler(csvService)

	type fields struct {
		endingBalance float64
		summary       *model.Summary
	}

	type want struct {
		statusCode int
		err        error
	}

	tests := []struct {
		name         string
		fields       fields
		expectations func(fields fields)
		want         want
	}{
		{
			name: "ok",
			fields: fields{
				endingBalance: 10,
			},
			expectations: func(fields fields) {
				handler.service.(*Mockservice).
					EXPECT().
					ProcessCsv(gomock.Any()).
					Return(fields.summary, nil)
			},
			want: want{
				statusCode: http.StatusOK,
				err:        nil,
			},
		},
		{
			name: "fail",
			fields: fields{
				endingBalance: 0,
			},
			expectations: func(fields fields) {
				handler.service.(*Mockservice).
					EXPECT().
					ProcessCsv(gomock.Any()).
					Return(fields.summary, errors.New("fail"))
			},
			want: want{
				statusCode: http.StatusInternalServerError,
				err:        errors.New("fail"),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.expectations(tc.fields)
			got, err := handler.LambdaEvent()
			assert.Equal(t, tc.want.statusCode, got.StatusCode)
			if err != nil {
				assert.Equal(t, tc.want.err.Error(), err.Error())
			}
		})
	}
}
