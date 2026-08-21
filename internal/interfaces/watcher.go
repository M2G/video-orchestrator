package interfaces

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"video-orchestrator/internal/domain"

	"github.com/sirupsen/logrus"
)

type HTTPServer struct {
	repo       domain.Repository
	uploadDir  string
	streamsDir string
	log        *logrus.Logger
	server     *http.Server
}

func NewHTTPServer(
	repo domain.Repository,
	uploadDir string,
	streamsDir string,
	log *logrus.Logger,
) *HTTPServer {
	h := &HTTPServer{
		repo:       repo,
		uploadDir:  uploadDir,
		streamsDir: streamsDir,
		log:        log,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", h.handleUpload)
	mux.HandleFunc("GET /jobs/{id}", h.handleGetJob)
	mux.HandleFunc("GET /streams/video/{id}/{filename}", h.handleServeFile)

	h.server = &http.Server{
		Handler: mux,
	}

	return h
}

func (h *HTTPServer) Start(addr string) error {
	h.server.Addr = addr
	h.log.WithField("addr", addr).Info("http_server_started")
	return h.server.ListenAndServe()
}

func (h *HTTPServer) Stop(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

func (h *HTTPServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid content-type")
		return
	}

	boundary := params["boundary"]
	if boundary == "" {
		h.sendError(w, http.StatusBadRequest, "missing boundary")
		return
	}

	mr := multipart.NewReader(r.Body, boundary)
	part, err := mr.NextPart()
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid multipart")
		return
	}
	defer part.Close()

	filename := part.FileName()
	if !strings.HasSuffix(filename, ".mp4") {
		h.sendError(w, http.StatusBadRequest, "only MP4 files are accepted")
		return
	}

	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to create upload dir")
		return
	}

	destPath := filepath.Join(h.uploadDir, filename)
	f, err := os.Create(destPath)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to create file")
		return
	}
	defer f.Close()

	const maxSize = 2 << 30 // 2GB
	if _, err := io.Copy(f, io.LimitReader(part, maxSize)); err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to write file")
		return
	}

	job, err := h.repo.CreateJob(r.Context(), destPath)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to create job")
		return
	}

	h.log.WithField("job_id", job.ID).Info("job_created")
	h.sendJSON(w, http.StatusCreated, map[string]any{
		"job_id":   job.ID,
		"filename": job.Filename,
	})
}

func (h *HTTPServer) handleGetJob(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid job id")
		return
	}

	job, err := h.repo.GetJobByID(r.Context(), id)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "job not found")
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]any{
		"job_id":   job.ID,
		"filename": job.Filename,
		"status":   job.Status,
	})
}

func (h *HTTPServer) handleServeFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")

	filePath := filepath.Join(h.streamsDir, id, filename)

	ext := filepath.Ext(filename)
	switch ext {
	case ".m3u8":
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	case ".ts":
		w.Header().Set("Content-Type", "video/mp2t")
	default:
		h.sendError(w, http.StatusForbidden, "file type not allowed")
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000")
	http.ServeFile(w, r, filePath)
}

func (h *HTTPServer) sendJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func (h *HTTPServer) sendError(w http.ResponseWriter, status int, msg string) {
	h.sendJSON(w, status, map[string]string{"error": msg})
}
