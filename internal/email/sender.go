package email

import (
	"bytes"
	"context"
	_ "embed"
	"html/template"
	"stori-challenge/internal/integrations/aws/ses"
	"stori-challenge/internal/model"
	"strconv"
)

const (
	noreplyEmail    = "benitezjulietasol@gmail.com"
	SubjectEmail    = "CSV summary account information"
	prefixMonthRow  = "<tr><td align=\"center\" width=\"20%\" style=\"margin: 10px 0 0;font-family: Archivo, Arial, Helvetica, sans-serif;font-style: normal;font-weight: normal;font-size: 14px;line-height: 22px;color: #070715;\">"
	suffixMonthRow  = "</td>"
	prefixAmountRow = "<td align=\"center\" width=\"20%\" style=\"margin: 10px 0 0;font-family: Archivo, Arial, Helvetica, sans-serif;font-style: normal;font-weight: normal;font-size: 14px;line-height: 22px;color: #070715;\">"
	suffixAmountRow = "</td></tr>"
)

//go:embed template.html
var TemplateSummary string

type emailService interface {
	SendEmail(context.Context, ses.SendEmailParams) error
}

type Service struct {
	emailService emailService
}

func NewService(emailService emailService) *Service {
	return &Service{emailService}
}

// RenderTemplate Helper function to write HTML templates based in payload info
func renderTemplate(tplStr string, payload interface{}) (string, error) {
	t := template.New("some_tpl")
	tpl, err := t.Parse(tplStr)

	if err != nil {
		//logger.Errorf("failed to parse template %s", err.Error())
		return "", err
	}

	var buf bytes.Buffer

	if err := tpl.Execute(&buf, payload); err != nil {
		//logger.Errorf("failed to parse template %s", err.Error())
		return "", err
	}

	return buf.String(), nil
}

func (s *Service) SendEmail(ctx context.Context, params model.EmailParams) error {
	html, err := renderTemplate(params.Template, params.Payload)
	if err != nil {
		return err
	}

	return s.emailService.SendEmail(
		ctx,
		ses.SendEmailParams{
			From:    noreplyEmail,
			To:      params.To,
			Subject: params.Subject,
			Html:    html,
		},
	)
}

func GenerateMonthlySummary(transactions map[string]int) template.HTML {
	var html string
	for month, amount := range transactions {
		monthRow := prefixMonthRow + month + suffixMonthRow + prefixAmountRow + strconv.Itoa(amount) + suffixAmountRow
		html = html + monthRow
	}

	return template.HTML(html)
}
