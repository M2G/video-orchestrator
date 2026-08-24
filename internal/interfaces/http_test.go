package interfaces_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"video-orchestrator/internal/domain"
	"video-orchestrator/internal/interfaces"
)

var (
	testServer     *httptest.Server
	testUploadDir  string
	testStreamsDir string
	testRepo       *mockRepoHTTP
)

func TestMain(m *testing.M) {
	testUploadDir, _ = os.MkdirTemp("", "upload-*")
	testStreamsDir, _ = os.MkdirTemp("", "streams-*")
	defer os.RemoveAll(testUploadDir)
	defer os.RemoveAll(testStreamsDir)

	testRepo = defaultRepoHTTP()
	interfaces.NewHTTPServer(testRepo, testUploadDir, testStreamsDir, noopLogger())
	testServer = httptest.NewServer(http.DefaultServeMux)
	defer testServer.Close()

	m.Run()
}

func TestHandleUpload_Success(t *testing.T) {
	testRepo.createJobFn = func(_ context.Context, filename string) (domain.VideoJob, error) {
		return domain.VideoJob{ID: 1, Filename: filename}, nil
	}

	body, contentType := multipartBody(t, "test.mp4", []byte("fake mp4 content"))

	resp, err := http.Post(testServer.URL+"/jobs", contentType, body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if _, ok := result["job_id"]; !ok {
		t.Error("expected job_id in response")
	}
}

func TestHandleUpload_InvalidContentType(t *testing.T) {
	resp, err := http.Post(testServer.URL+"/jobs", "application/json", bytes.NewBufferString("data"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleUpload_NotMP4(t *testing.T) {
	body, contentType := multipartBody(t, "test.avi", []byte("fake avi content"))

	resp, err := http.Post(testServer.URL+"/jobs", contentType, body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleGetJob_Success(t *testing.T) {
	testRepo.getJobByIDFn = func(_ context.Context, id int64) (domain.VideoJobDetails, error) {
		return domain.VideoJobDetails{
			VideoJob: domain.VideoJob{ID: id, Filename: "test.mp4"},
			Status:   domain.StatusPending,
		}, nil
	}

	resp, err := http.Get(testServer.URL + "/jobs/1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result["status"] != string(domain.StatusPending) {
		t.Errorf("expected status PENDING, got %v", result["status"])
	}
}

func TestHandleGetJob_NotFound(t *testing.T) {
	testRepo.getJobByIDFn = func(_ context.Context, id int64) (domain.VideoJobDetails, error) {
		return domain.VideoJobDetails{}, fmt.Errorf("job not found")
	}

	resp, err := http.Get(testServer.URL + "/jobs/99")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandleServeFile_M3U8(t *testing.T) {
	videoDir := filepath.Join(testStreamsDir, "1")
	if err := os.MkdirAll(videoDir, 0755); err != nil {
		t.Fatal(err)
	}
	m3u8Content := []byte("#EXTM3U\n#EXT-X-VERSION:3\n")
	if err := os.WriteFile(filepath.Join(videoDir, "index.m3u8"), m3u8Content, 0644); err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(testServer.URL + "/streams/video/1/index.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "application/vnd.apple.mpegurl" {
		t.Errorf("expected m3u8 content-type, got %s", ct)
	}

	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("#EXTM3U")) {
		t.Error("expected m3u8 content in response")
	}
}

func TestHandleServeFile_NotFound(t *testing.T) {
	resp, err := http.Get(testServer.URL + "/streams/video/99/index.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode) // StatusDate -> StatusCode
	}
}

func TestHandleServeFile_ForbiddenExtension(t *testing.T) {
	resp, err := http.Get(testServer.URL + "/streams/video/1/video.mp4")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}
