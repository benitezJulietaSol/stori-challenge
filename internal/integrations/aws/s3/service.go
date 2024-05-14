package s3

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

const (
	REGION    = "awsRegion"
	KEY       = "awsKey"
	SECRET    = "awsSecret"
	BUCKET    = "awsBucket"
	BUCKETKEY = "awsBucketKey"
)

type S3Service struct {
	region     string
	client     *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	bucket     string
	key        string
}

func NewS3Service(config map[string]string) *S3Service {
	region := config[REGION]
	sess, err := session.NewSession(&aws.Config{
		Region: &(region),
		Credentials: credentials.NewStaticCredentials(
			config[KEY],
			config[SECRET],
			"",
		),
	})

	if err != nil {
		panic(errors.Wrap(err, "failed to init aws session for ecs"))
	}

	return &S3Service{
		region:     region,
		client:     s3.New(sess),
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
		bucket:     config[BUCKET],
		key:        config[BUCKETKEY],
	}
}

func (s *S3Service) ReadFile(ctx context.Context) ([]byte, error) {
	buf := aws.NewWriteAtBuffer([]byte{})
	_, err := s.downloader.DownloadWithContext(ctx, buf, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &s.key,
	})

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
