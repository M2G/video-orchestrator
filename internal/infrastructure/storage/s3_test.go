package storage

import (
	"context"
	"os"
	"testing"
)

func TestS3Upload(t *testing.T) {

	StartFakeS3()

	ctx := context.Background()

	s := NewFakeS3("videos")
	s.CreateBucket(ctx)

	tmpFile := "test.txt"
	os.WriteFile(tmpFile, []byte("hello"), 0644)
	defer os.Remove(tmpFile)

	err := s.Upload(ctx, tmpFile, "test.txt")
	if err != nil {
		t.Fatal(err)
	}
}
