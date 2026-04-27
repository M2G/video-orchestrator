package storage

import (
	"log"
	"net/http"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
)

func StartFakeS3() {
	backend := s3mem.New()
	faker := gofakes3.New(backend)

	server := &http.Server{
		Addr:    ":9000",
		Handler: faker.Server(),
	}

	log.Println("Fake S3 running on :9000")

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
}
