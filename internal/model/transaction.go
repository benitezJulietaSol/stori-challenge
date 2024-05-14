package model

import (
	"html/template"
	"time"
)

const (
	DEBIT  = "credit"
	CREDIT = "debit"
)

type Transaction struct {
	ID     float64   `json:"id" pg:",use_zero"`
	Amount float64   `json:"amount"`
	Date   time.Time `json:"date"`
}

type Summary struct {
	Debit          []Transaction `json:"debit"`
	Credit         []Transaction `json:"credit"`
	RunningBalance float64       `json:"balance"`
}

func (s *Summary) GetAverage(trx string) float64 {
	var (
		total        float64 = 0
		transactions []Transaction
		length       int
	)

	if trx == DEBIT {
		length = len(s.Debit)
		if length == 0 {
			return 0
		}
		transactions = s.Debit
	} else {
		length = len(s.Credit)
		if length == 0 {
			return 0
		}
		transactions = s.Credit
	}

	for i := 0; i < length; i++ {
		total += transactions[i].Amount
	}

	return total / float64(length)
}

type EmailParams struct {
	To                []string
	Subject, Template string
	Payload           interface{}
}

type Data struct {
	Name                string
	EndingBalance       string
	DebitAmount         string
	CreditAmount        string
	MonthlyTransactions template.HTML
}
