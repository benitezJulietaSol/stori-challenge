package transaction

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path/filepath"
	"stori-challenge/internal/model"
	"testing"
	"time"
)

func GetBytesFile(filePath string) ([]byte, error) {
	path, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(path)
}

func toDate(record string) time.Time {
	date, err := time.Parse("1/2", record)
	if err != nil {
		fmt.Println(err.Error())
	}

	return date
}

func TestProcessCsv(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txsRepository := NewMockrepository(ctrl)
	sesService := NewMockemailService(ctrl)
	bucketService := NewMocks3Service(ctrl)
	procService := NewService(
		sesService,
		bucketService,
		txsRepository,
	)

	requestRaw, err := GetBytesFile("input/transactions.csv")
	if err != nil {
		fmt.Println(err.Error())
	}

	type fields struct {
		transactions []model.Transaction
	}

	type want struct {
		endingBalance float64
		summary       *model.Summary
		err           error
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
				[]model.Transaction{
					{0, +60.5, toDate("7/15")},
					{3, +10, toDate("8/13")},
					{1, -10.3, toDate("7/28")},
					{2, -20.46, toDate("8/2")},
				},
			},
			expectations: func(fields fields) {
				procService.bucket.(*Mocks3Service).
					EXPECT().
					ReadFile(gomock.Any()).
					Return(requestRaw, nil)
				procService.email.(*MockemailService).
					EXPECT().
					SendEmail(gomock.Any(), gomock.Any()).
					Return(nil)
				procService.repository.(*Mockrepository).
					EXPECT().
					InsertTransactions(gomock.Any(), fields.transactions).
					Return(nil)
			},
			want: want{
				endingBalance: 39.74,
				summary: &model.Summary{
					Debit: []model.Transaction{
						{ID: 0, Amount: 60.5, Date: time.Date(0, time.July, 15, 0, 0, 0, 0, time.UTC)},
						{ID: 3, Amount: 10, Date: time.Date(0, time.August, 13, 0, 0, 0, 0, time.UTC)}},
					Credit: []model.Transaction{
						{ID: 1, Amount: -10.3, Date: time.Date(0, time.July, 28, 0, 0, 0, 0, time.UTC)},
						{ID: 2, Amount: -20.46, Date: time.Date(0, time.August, 2, 0, 0, 0, 0, time.UTC)}},
					RunningBalance: 39.74,
				},
				err: nil,
			},
		},
		{
			name: "fail_insert",
			fields: fields{
				[]model.Transaction{
					{0, +60.5, toDate("7/15")},
					{3, +10, toDate("8/13")},
					{1, -10.3, toDate("7/28")},
					{2, -20.46, toDate("8/2")},
				},
			},
			expectations: func(fields fields) {
				procService.bucket.(*Mocks3Service).
					EXPECT().
					ReadFile(gomock.Any()).
					Return(requestRaw, nil)
				procService.email.(*MockemailService).
					EXPECT().
					SendEmail(gomock.Any(), gomock.Any()).
					Return(nil)
				procService.repository.(*Mockrepository).
					EXPECT().
					InsertTransactions(gomock.Any(), fields.transactions).
					Return(errors.New("fail"))
			},
			want: want{
				endingBalance: 39.74,
				summary: &model.Summary{
					Debit: []model.Transaction{
						{ID: 0, Amount: 60.5, Date: time.Date(0, time.July, 15, 0, 0, 0, 0, time.UTC)},
						{ID: 3, Amount: 10, Date: time.Date(0, time.August, 13, 0, 0, 0, 0, time.UTC)}},
					Credit: []model.Transaction{
						{ID: 1, Amount: -10.3, Date: time.Date(0, time.July, 28, 0, 0, 0, 0, time.UTC)},
						{ID: 2, Amount: -20.46, Date: time.Date(0, time.August, 2, 0, 0, 0, 0, time.UTC)}},
					RunningBalance: 39.74,
				},
				err: nil,
			},
		},
		{
			name:   "error_records",
			fields: fields{},
			expectations: func(fields fields) {
				procService.bucket.(*Mocks3Service).
					EXPECT().
					ReadFile(gomock.Any()).
					Return(requestRaw, errors.New("fail"))
			},
			want: want{
				endingBalance: 0,
				summary:       &model.Summary{},
				err:           errors.New("fail"),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.expectations(tc.fields)
			summary, balance, err := procService.ProcessCsv(context.Background())

			assert.EqualValues(t, tc.want.summary, summary)
			assert.EqualValues(t, tc.want.endingBalance, balance)
			if err != nil {
				assert.Equal(t, tc.want.err.Error(), err.Error())
			}
		})
	}
}
