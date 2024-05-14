package ses

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	ses         *sesv2.SESV2
	defaultFrom string
}

type SendEmailParams struct {
	From    string
	To      []string
	Subject string
	Text    string
	Html    string
}

func NewService(config map[string]string) *Service {
	region := config["awsRegion"]
	sess, err := session.NewSession(&aws.Config{
		Region: &(region),
		Credentials: credentials.NewStaticCredentials(
			config["awsKey"],
			config["awsSecret"],
			"",
		),
	})

	if err != nil {
		panic(errors.Wrap(err, "failed to init aws session for ses"))
	}

	return &Service{ses: sesv2.New(sess), defaultFrom: config["awsSesFrom"]}
}

func (s *Service) SendEmail(ctx context.Context, details SendEmailParams) error {
	charset := "UTF-8"
	body := sesv2.Body{}
	to := make([]*string, len(details.To))

	for i := range details.To {
		to[i] = &details.To[i]
	}

	if len(details.Html) > 0 {
		body.Html = &sesv2.Content{
			Charset: &charset,
			Data:    &details.Html,
		}
	} else {
		body.Text = &sesv2.Content{
			Charset: &charset,
			Data:    &details.Text,
		}
	}

	from := details.From

	if from == "" {
		from = s.defaultFrom
	}

	_, err := s.ses.SendEmail(&sesv2.SendEmailInput{
		FromEmailAddress: &from,
		Destination: &sesv2.Destination{
			ToAddresses: to,
		},
		Content: &sesv2.EmailContent{
			Simple: &sesv2.Message{
				Body: &body,
				Subject: &sesv2.Content{
					Charset: &charset,
					Data:    &details.Subject,
				},
			},
		},
	})

	log.WithContext(ctx).Infof("Sending email to: %v", to)

	if err != nil {
		log.WithContext(ctx).Errorf("failed to send email via aws ses %s", err.Error())
		return err
	}

	return nil
}
