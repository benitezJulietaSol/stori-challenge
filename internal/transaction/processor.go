package transaction

import (
	"awesomeProject2/internal/email"
	model "awesomeProject2/internal/model"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=transaction

const (
	columnIndexId = iota
	columnIndexDate
	columnIndexAmount
)

type emailService interface {
	SendEmail(context.Context, model.EmailParams) error
}

type s3Service interface {
	ReadFile(ctx context.Context) ([]byte, error)
}

type repository interface {
	InsertTransactions(context.Context, []model.Transaction) error
}

type Service struct {
	repository repository
	email      emailService
	bucket     s3Service
}

func NewService(emailService emailService, s3Service s3Service, repo repository) *Service {
	return &Service{
		email:      emailService,
		bucket:     s3Service,
		repository: repo,
	}
}

func (s *Service) getTransactions(ctx context.Context) ([]byte, error) {
	return s.bucket.ReadFile(ctx)
}

func (s *Service) getRecordsFromFile(ctx context.Context) ([][]string, error) {
	file, err := s.getTransactions(ctx)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to get transactions %s", err.Error())
		return nil, err
	}
	reader := bytes.NewReader(file)
	csvReader := csv.NewReader(reader)
	csvReader.Comma = ','

	return csvReader.ReadAll()
}

func (s *Service) ProcessCsv(ctx context.Context) (
	summary *model.Summary,
	runningBalance float64,
	err error,
) {
	monthlyGrouping := make(map[string]int)
	summary = &model.Summary{}
	emails := []string{"benitezjulietasol@gmail.com"}

	defer func(emails []string) {
		if err == nil {
			err = s.email.SendEmail(ctx, model.EmailParams{
				To:       emails,
				Subject:  email.SubjectEmail,
				Template: email.TemplateSummary,
				Payload: model.Data{
					Name:                fmt.Sprintf("%s", emails),
					EndingBalance:       fmt.Sprintf("%f", runningBalance),
					DebitAmount:         fmt.Sprintf("%.2f", summary.GetAverage(model.DEBIT)),
					CreditAmount:        fmt.Sprintf("%.2f", summary.GetAverage(model.CREDIT)),
					MonthlyTransactions: email.GenerateMonthlySummary(monthlyGrouping),
				},
			})
			if err != nil {
				log.WithContext(ctx).Fatal(err)
				return
			}
		}
	}(emails)

	const bitSize = 64
	var (
		headerIndex = 0
	)

	records, err := s.getRecordsFromFile(ctx)
	if err != nil {
		log.WithContext(ctx).Fatal(err)
	}
	fmt.Println(records)

	for i, record := range records {
		if i <= headerIndex {
			continue
		}

		log.WithContext(ctx).Infof("Parsing row at index %d: %v", i, record)

		id, err := strconv.ParseFloat(record[columnIndexId], bitSize)
		if err != nil {
			log.WithContext(ctx).Errorf("failed to parse date %s at row %d", record[columnIndexDate], i+1)
			return nil, 0, err
		}

		// Parse date
		date, err := time.Parse("1/2", record[columnIndexDate])
		if err != nil {
			log.WithContext(ctx).Errorf("failed to parse date %s at row %d", record[columnIndexDate], i+1)
			return nil, 0, err
		}

		log.WithContext(ctx).Infof("Parsed date: %v", date)

		// Parse amount. debit is negative and credit is positive
		amount, err := strconv.ParseFloat(record[columnIndexAmount], bitSize)
		if err != nil {
			log.WithContext(ctx).Errorf("failed to parse amount %s at row %d", record[columnIndexAmount], i+1)
			return nil, 0, err
		}

		log.WithContext(ctx).Infof("Parsed amount: %v", amount)

		//
		monthlyGrouping[date.Month().String()] = monthlyGrouping[date.Month().String()] + 1
		//
		transaction := model.Transaction{
			ID:     id,
			Amount: amount,
			Date:   date,
		}

		if amount > 0 {
			summary.Debit = append(summary.Debit, transaction)
		} else {
			summary.Credit = append(summary.Credit, transaction)
		}

		runningBalance += amount
	}

	// insertar todas las transacciones
	go func() {
		transactions := summary.Debit
		transactions = append(transactions, summary.Credit...)
		err := s.repository.InsertTransactions(ctx, transactions)
		if err != nil {
			log.WithContext(ctx).Error("insert transactions")
		}
	}()

	return summary, runningBalance, nil
}
