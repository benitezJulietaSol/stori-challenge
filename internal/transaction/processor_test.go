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

	type want struct {
		summary *model.Summary
		err     error
	}

	tests := []struct {
		name         string
		expectations func()
		want         want
	}{
		{
			name: "error_records",
			expectations: func() {
				procService.bucket.(*Mocks3Service).
					EXPECT().
					ReadFile(gomock.Any()).
					Return(requestRaw, errors.New("fail"))
			},
			want: want{
				summary: &model.Summary{},
				err:     errors.New("fail"),
			},
		},
		{
			name: "ok",
			expectations: func() {
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
					InsertTransactions(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			},
			want: want{
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
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.expectations()
			summary, err := procService.ProcessCsv(context.Background())

			assert.ElementsMatch(t, tc.want.summary.Debit, summary.Debit)
			assert.ElementsMatch(t, tc.want.summary.Credit, summary.Credit)
			if err != nil {
				assert.Equal(t, tc.want.err.Error(), err.Error())
			}
		})
	}
}
