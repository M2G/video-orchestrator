package interfaces_test

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"testing"

	"video-orchestrator/internal/domain"

	"github.com/sirupsen/logrus"
)

type mockRepoHTTP struct {
	createJobFn  func(ctx context.Context, filename string) (domain.VideoJob, error)
	getJobByIDFn func(ctx context.Context, id int64) (domain.VideoJobDetails, error)
	existsFn     func(ctx context.Context, id int64) (bool, error)
	markDoneFn   func(ctx context.Context, id int64) error
}

func (m *mockRepoHTTP) CreateJob(ctx context.Context, filename string) (domain.VideoJob, error) {
	return m.createJobFn(ctx, filename)
}

func (m *mockRepoHTTP) GetJobByID(ctx context.Context, id int64) (domain.VideoJobDetails, error) {
	return m.getJobByIDFn(ctx, id)
}

func (m *mockRepoHTTP) Exists(ctx context.Context, id int64) (bool, error) {
	return m.existsFn(ctx, id)
}

func (m *mockRepoHTTP) MarkDone(ctx context.Context, id int64) error {
	return m.markDoneFn(ctx, id)
}

func defaultRepoHTTP() *mockRepoHTTP {
	return &mockRepoHTTP{
		createJobFn: func(_ context.Context, filename string) (domain.VideoJob, error) {
			return domain.VideoJob{ID: 1, Filename: filename}, nil
		},
		getJobByIDFn: func(_ context.Context, id int64) (domain.VideoJobDetails, error) {
			return domain.VideoJobDetails{
				VideoJob: domain.VideoJob{ID: id, Filename: "test.mp4"},
				Status:   domain.StatusPending,
			}, nil
		},
		existsFn:   func(_ context.Context, id int64) (bool, error) { return true, nil },
		markDoneFn: func(_ context.Context, id int64) error { return nil },
	}
}

func multipartBody(t *testing.T, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	part.Write(content)
	writer.Close()
	return body, writer.FormDataContentType()
}

func noopLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return log
}
