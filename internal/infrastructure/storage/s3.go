package storage

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client *s3.Client
	bucket string
}

func NewFakeS3(bucket string) *S3Storage {

	cfg, _ := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(
				func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL:           "http://localhost:9000",
						SigningRegion: "us-east-1",
					}, nil
				},
			),
		),
	)

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3Storage{
		client: client,
		bucket: bucket,
	}
}

func (s *S3Storage) Upload(ctx context.Context, path, key string) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
		Body:   file,
	})

	return err
}

func (s *S3Storage) CreateBucket(ctx context.Context) {
	s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: &s.bucket,
	})
}
