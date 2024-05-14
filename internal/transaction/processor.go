package transaction

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	log "github.com/sirupsen/logrus"
	"stori-challenge/internal/email"
	model "stori-challenge/internal/model"
	"strconv"
	"sync"
	"time"
)

//go:generate mockgen -source=processor.go -destination=processor_mock.go -package=transaction

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
	const numWorkers = 5
	var (
		monthlyGrouping = make(map[string]int)
		emails          = []string{"benitezjulietasol@gmail.com"}
		wg              sync.WaitGroup
		recordChan      = make(chan []string)
		errChan         = make(chan error, 1)
	)

	summary = &model.Summary{}

	defer func(emails []string) {
		if err == nil {
			err = s.email.SendEmail(ctx, model.EmailParams{
				To:       emails,
				Subject:  email.SubjectEmail,
				Template: email.TemplateSummary,
				Payload: model.Data{
					Name:                fmt.Sprintf("%s", emails),
					EndingBalance:       fmt.Sprintf("%.2f", runningBalance),
					DebitAmount:         fmt.Sprintf("%.2f", summary.GetAverage(model.DEBIT)),
					CreditAmount:        fmt.Sprintf("%.2f", summary.GetAverage(model.CREDIT)),
					MonthlyTransactions: email.GenerateMonthlySummary(monthlyGrouping),
				},
			})
			if err != nil {
				log.WithContext(ctx).Error(err)
				return
			}
		}
	}(emails)

	records, err := s.getRecordsFromFile(ctx)
	if err != nil {
		log.WithContext(ctx).Error(err)
		return
	}

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for record := range recordChan {
				err := processRecord(
					ctx,
					record,
					summary,
					monthlyGrouping,
					&runningBalance,
				)
				if err != nil {
					errChan <- err
					return
				}
			}
		}()
	}

	// sync records to workers
	headerIndex := 0
	for i, record := range records {
		if i <= headerIndex {
			continue
		}
		select {
		case recordChan <- record:
		case err := <-errChan:
			close(recordChan)
			return nil, 0, err
		}
	}
	close(recordChan)
	wg.Wait()

	close(errChan)
	for e := range errChan {
		if e != nil {
			return nil, 0, e
		}
	}

	go func() {
		transactions := summary.Debit
		transactions = append(transactions, summary.Credit...)
		err := s.repository.InsertTransactions(ctx, transactions)
		if err != nil {
			log.WithContext(ctx).Error("Fail to insert transactions, error: %s", err.Error())
		}
	}()

	summary.RunningBalance = runningBalance
	return summary, runningBalance, nil
}

func processRecord(ctx context.Context, record []string, summary *model.Summary, monthlyGrouping map[string]int, runningBalance *float64) error {
	const bitSize = 64

	log.WithContext(ctx).Infof("Parsing row: %v", record)

	id, err := strconv.ParseFloat(record[columnIndexId], bitSize)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to parse date %s:", record[columnIndexDate])
		return err
	}

	// Parse date
	date, err := time.Parse("1/2", record[columnIndexDate])
	if err != nil {
		log.WithContext(ctx).Errorf("failed to parse date %s", record[columnIndexDate])
		return err
	}

	log.WithContext(ctx).Infof("Parsed date: %v", date)

	// Parse amount. debit is negative and credit is positive
	amount, err := strconv.ParseFloat(record[columnIndexAmount], bitSize)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to parse amount %s", record[columnIndexAmount])
		return err
	}

	log.WithContext(ctx).Infof("Parsed amount: %v", amount)

	monthlyGrouping[date.Month().String()] = monthlyGrouping[date.Month().String()] + 1

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

	*runningBalance += amount

	return nil
}
